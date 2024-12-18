import sys
from yt_dlp import YoutubeDL
import os

def download_youtube_video_and_audio(video_url, video_path, audio_path):
    try:
        # Ensure the directories for the specified paths exist
        # os.makedirs(os.path.dirname(video_path), exist_ok=True)
        # os.makedirs(os.path.dirname(audio_path), exist_ok=True)

        # Download the highest quality video as MP4 to the specified video_path
        video_options = {
            'format': 'bestvideo',  # Get the highest quality video
            'noplaylist': True,     # Avoid downloading playlists
            'quiet': False,         # Show detailed logs
            'outtmpl': video_path,  # Save video to the exact specified path
            'merge_output_format': None,  # Don't append .mp4
        }

        print(f"Downloading the highest quality video to {video_path}...")
        with YoutubeDL(video_options) as ydl:
            ydl.download([video_url])


        # Download the highest quality audio as MP4 to the specified audio_path
        audio_options = {
            'format': 'bestaudio',  # Get the highest quality audio
            'noplaylist': True,     # Avoid downloading playlists
            'quiet': False,         # Show detailed logs
            'outtmpl': audio_path.split(".")[0] + ".mp4",  # Save audio to the exact specified path
            'merge_output_format': None,  # Don't append .mp4
            'postprocessors': [
                {'key': 'FFmpegVideoConvertor', 'preferedformat': 'mp4'}  # Convert audio to MP4
            ]
        }

        print(f"Downloading the highest quality audio to {audio_path}...")
        with YoutubeDL(audio_options) as ydl:
            ydl.download([video_url])

        print("Download completed successfully!")
    except Exception as e:
        print(f"Error during download: {e}")
        sys.exit(1)

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("Usage: python youtube_download.py <video_url> <video_path> <audio_path>")
        sys.exit(1)

    video_url = sys.argv[1]
    video_path = sys.argv[2]
    audio_path = sys.argv[3]

    download_youtube_video_and_audio(video_url, video_path, audio_path)
