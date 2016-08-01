#!/bin/sh

set -e
go build
chmod 644 index.js
chmod 755 optionalcloud
zip -r lambda.zip index.js optionalcloud
