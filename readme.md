## Website Monitor AWS Lambda Slack App

An aws lambda function to monitor websites status.
It can be triggered by a CloudWatch event (cron every 5 minutes suggested).
Requires 3 parameters `CSV_URL` (csv file with list of websites to parse), `WEBHOOK_URL` (from slack incoming webhooks) and `CHANNEL` (the slack channel to use).
These can be provided either by env variables through the lambda config on AWS or as raw json input from CloudWatch event rule.

When more than one website has issues, a new message will be posted into the specified slack channel with the list of websites with errors

<p align="center">
  <img src="http://ns3003669.ip-37-187-17.eu/img/monitor_lambda_slack_thumb.png"/>
</p>

