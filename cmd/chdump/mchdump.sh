#!/bin/bash

# Print help if no arguments provided
if [ $# -eq 0 ]; then
    echo "Usage: $(basename $0) [options] [hosts...]"
    echo
    echo "Options:"
    echo "  -u <username>    Database username (default: 'default')"
    echo "  -p <password>    Database password (will prompt if not provided)" 
    echo "  -d <database>    Database name (required)"
    echo "  -o <file>        Output file name (default: schemas.tar.gz)"
    echo
    echo "Arguments:"
    echo "  hosts            Space-separated list of database hosts"
    echo "                   (default: localhost)"
    echo
    echo "Example:"
    echo "  $(basename $0) -u admin -d mydb host1 host2"
    exit 1
fi


# Parse command line arguments
username="default"
while getopts "u:p:d:o:" opt; do
  case $opt in
    u)
      username="$OPTARG"
      ;;
    p)
      password="$OPTARG"
      ;;
    d)
      database="$OPTARG"
      ;;
    o)
      output_file="$OPTARG"
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
    :)
      echo "Option -$OPTARG requires an argument." >&2
      exit 1
      ;;
  esac
done

# Store remaining arguments as hosts
shift $((OPTIND-1))
hosts="$@"


if [ -z "$database" ]; then
  echo "Error: database argument is required" >&2
  exit 1
fi
if [ -z "$hosts" ]; then
  hosts=("localhost")
fi
if [ -z "$output_file" ]; then
  output_file="schemas-for-${database}-at-$(date '+%Y-%m-%d_%H-%M-%S').tar.gz"
fi

if [ -z "$password" ]; then
  echo -n "Enter password for $username: "
  read -s password
  echo
fi

# Check if output file already exists
if [ -f "$output_file" ]; then
  echo "Error: output file '$output_file' already exists" >&2
  exit 1
fi

# Create temporary directory
temp_dir=$(mktemp -d)
if [ $? -ne 0 ]; then
  echo "Failed to create temporary directory" >&2
  exit 1
fi

# Clean up temp dir on exit
trap 'rm -rf "$temp_dir"' EXIT

for host in $hosts; do
    echo Getting $database schema from $host
    url="tcp://$host:9000/$database?username=$username&password=$password"
    chdump "$url" > "$temp_dir/$host.sql"
done

# Create tar archive from temp directory contents
tar -czf $output_file -C $temp_dir .

