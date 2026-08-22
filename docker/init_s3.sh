#!/bin/bash
awslocal s3 mb s3://local-bucket
awslocal s3api put-public-access-block --bucket local-bucket --public-access-block BlockPublicAcls=false,IgnorePublicAcls=false,BlockPublicPolicy=false,RestrictPublicBuckets=false
cat << 'EOF' > /tmp/policy.json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "PublicReadGetObject",
            "Effect": "Allow",
            "Principal": "*",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::local-bucket/*"
        }
    ]
}
EOF
awslocal s3api put-bucket-policy --bucket local-bucket --policy file:///tmp/policy.json