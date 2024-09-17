# Qovery AI Migration Web Backend

This project use the Golang Qovery Migration AI Agent library and expose it via a REST API.

## Environment Variables

The following environment variables are required to run the application:

| Environment Variable   | Description                              | Required           |
|------------------------|------------------------------------------|--------------------|
| `CLAUDE_API_KEY`       | Claude AI API key                        | Yes                |
| `HEROKU_API_KEY`       | Heroku API key                           | Yes if you used it |
| `GITHUB_TOKEN`         | GitHub token to avoid being rate limited | No                 |
| `S3_BUCKET`            | S3 bucket for storing migration files    | No                 |
| `S3_REGION`            | S3 region for the bucket                 | No                 |
| `S3_ACCESS_KEY`        | S3 access key for the bucket             | No                 |
| `S3_SECRET_ACCESS_KEY` | S3 secret key for the bucket             | No                 |

For S3 storage, ensure that the bucket is created and the access keys are configured properly.

### IAM Permissions

Make sure that the IAM user or role associated with the `S3_ACCESS_KEY` has the necessary permissions to access the specified S3 bucket.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject"
      ],
      "Resource": "arn:aws:s3:::your-bucket-name/*"
    }
  ]
}
```