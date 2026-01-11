#!/bin/bash

RABBITMQ_URL="https://rabbitmq-production-b72c.up.railway.app"
RABBITMQ_USER="woragis"
RABBITMQ_PASSWORD="woragis"
CREDENTIALS="$RABBITMQ_USER:$RABBITMQ_PASSWORD"

echo "Setting up RabbitMQ vhosts..."

for VHOST in jobs posts resume management; do
  echo "Creating vhost /$VHOST..."
  curl -k -X PUT \
    -u "$CREDENTIALS" \
    "$RABBITMQ_URL:15672/api/vhosts/%2F$VHOST" \
    -H "Content-Type: application/json" \
    -d '{}'
  echo ""
done

echo "Setting permissions..."
for VHOST in jobs posts resume management; do
  echo "Setting permissions for vhost /$VHOST..."
  curl -k -X PUT \
    -u "$CREDENTIALS" \
    "$RABBITMQ_URL:15672/api/permissions/%2F$VHOST/$RABBITMQ_USER" \
    -H "Content-Type: application/json" \
    -d '{"configure":".*","write":".*","read":".*"}'
  echo ""
done

echo "✓ RabbitMQ vhosts initialized"