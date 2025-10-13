#!/bin/sh
set -e

# Set template path for config-gen
export TEMPLATE_PATH=/etc/nginx/nginx.conf.tmpl

echo "Generating nginx configuration..."
/usr/local/bin/config-gen

# Validate nginx configuration
echo "Validating nginx configuration..."
nginx -t

echo "Starting nginx..."
exec "$@"
