#!/bin/bash
# integration_test.sh

export INTEGRATION_TEST=true
export POCKETBASE_URL=${POCKETBASE_URL:-http://localhost:8090}
export TEST_USER_EMAIL="valid@example.com"
export TEST_USER_PASSWORD="validpassword123"

go test -v ./internal/handlers -run TestLiveAuthentication