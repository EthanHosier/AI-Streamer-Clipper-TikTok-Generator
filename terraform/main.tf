provider "aws" {
  region = "us-east-1" # Replace with your preferred AWS region

  retry_mode  = "standard"
  max_retries = 3
}

resource "aws_s3_bucket" "clips_bucket" {
  bucket = "stream-clips-bucket"
}

resource "aws_s3_bucket_lifecycle_configuration" "clips_bucket_lifecycle" {
  bucket = aws_s3_bucket.clips_bucket.id

  rule {
    id     = "expire-clips"
    status = "Enabled"

    expiration {
      days = 7
    }
  }
}

resource "aws_s3_bucket_ownership_controls" "clips_bucket_ownership" {
  bucket = aws_s3_bucket.clips_bucket.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "clips_bucket_acl" {
  depends_on = [aws_s3_bucket_ownership_controls.clips_bucket_ownership]
  bucket     = aws_s3_bucket.clips_bucket.id
  acl        = "private"
}

resource "aws_cloudfront_origin_access_control" "oac" {
  name                              = "clips-bucket-oac"
  description                       = "Origin Access Control for Clips Bucket"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

resource "aws_cloudfront_distribution" "cdn" {
  origin {
    domain_name              = aws_s3_bucket.clips_bucket.bucket_regional_domain_name
    origin_id                = "S3-clips-bucket"
    origin_access_control_id = aws_cloudfront_origin_access_control.oac.id
  }

  enabled = true

  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "S3-clips-bucket"

    cache_policy_id          = "658327ea-f89d-4fab-a63d-7e88639e58f6" # CachingOptimized
    origin_request_policy_id = "88a5eaf4-2fd4-4709-b370-b4c650ea3fcf" # CORS-S3Origin

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  logging_config {
    include_cookies = false
    bucket          = aws_s3_bucket.logging_bucket.bucket_domain_name
    prefix          = "cloudfront-logs/"
  }
}

resource "aws_s3_bucket_policy" "clips_policy" {
  bucket = aws_s3_bucket.clips_bucket.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid    = "AllowCloudFrontAccess"
        Effect = "Allow"
        Principal = {
          Service = "cloudfront.amazonaws.com"
        }
        Action   = "s3:GetObject"
        Resource = "${aws_s3_bucket.clips_bucket.arn}/*"
        Condition = {
          StringEquals = {
            "AWS:SourceArn" = aws_cloudfront_distribution.cdn.arn
          }
        }
      }
    ]
  })
}

resource "aws_s3_bucket_cors_configuration" "clips_bucket_cors" {
  bucket = aws_s3_bucket.clips_bucket.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST"]
    allowed_origins = ["*"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}

resource "aws_s3_bucket_versioning" "clips_bucket_versioning" {
  bucket = aws_s3_bucket.clips_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket" "logging_bucket" {
  bucket = "stream-clips-logs-bucket"
}

resource "aws_s3_bucket_ownership_controls" "logging_bucket_ownership" {
  bucket = aws_s3_bucket.logging_bucket.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "logging_bucket_acl" {
  depends_on = [aws_s3_bucket_ownership_controls.logging_bucket_ownership]
  bucket     = aws_s3_bucket.logging_bucket.id
  acl        = "log-delivery-write"
}

output "cloudfront_domain_name" {
  description = "The domain name of the CloudFront distribution"
  value       = aws_cloudfront_distribution.cdn.domain_name
}
