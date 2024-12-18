terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Create a GCS bucket
resource "google_storage_bucket" "video_bucket" {
  name          = "${var.project_id}-video-uploads"
  location      = var.region
  force_destroy = true # Allows terraform destroy to remove bucket even if it contains objects

  uniform_bucket_level_access = true

  lifecycle_rule {
    condition {
      age = 1 # Delete files after 1 day
    }
    action {
      type = "Delete"
    }
  }
}

# Add public access to objects in the bucket
resource "google_storage_bucket_iam_member" "public_access" {
  bucket = google_storage_bucket.video_bucket.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

# Create a service account
resource "google_service_account" "video_uploader" {
  account_id   = "video-uploader"
  display_name = "Video Uploader Service Account"
  description  = "Service account for uploading videos to GCS"
}

# Grant the service account permissions to upload to the bucket
resource "google_storage_bucket_iam_member" "video_uploader" {
  bucket = google_storage_bucket.video_bucket.name
  role   = "roles/storage.objectUser"
  member = "serviceAccount:${google_service_account.video_uploader.email}"
}

# Create service account key
resource "google_service_account_key" "video_uploader_key" {
  service_account_id = google_service_account.video_uploader.name
}

# Output the bucket name and service account key
output "bucket_name" {
  value = google_storage_bucket.video_bucket.name
}

output "service_account_key" {
  value     = base64decode(google_service_account_key.video_uploader_key.private_key)
  sensitive = true
}
