env GOOS=linux GOARCH=amd64 go build -o /tmp/websiteMonitor && \
zip -j /tmp/websiteMonitor.zip /tmp/websiteMonitor && \
aws s3 cp /tmp/websiteMonitor.zip s3://{your-bucket} && \
aws lambda update-function-code  --function-name websiteMonitor --s3-bucket t{your-bucket} --s3-key websiteMonitor.zip
