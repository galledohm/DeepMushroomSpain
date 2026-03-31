# Collection Downloader

This directory contains the Go-based CSV image downloader.

## Files

- `download_images.go`: reads an iNaturalist CSV file, filters rows, and downloads images grouped by scientific name
- `go.mod`: local Go module for this directory

## Edit

When editing `download_images.go`, keep these points in mind:

- The script expects to be run from this directory because the CSV and image paths are relative to `src/collection/`
- CSV columns are resolved by header name, not by hard-coded column index
- Image filenames use the CSV `id` field
- Downloads are written to a temporary `.part` file first and only renamed after validation succeeds

## Go Module

From this directory, check the module file:

```powershell
Get-Content .\go.mod
```

If you need to refresh module metadata after future changes:

```powershell
go mod tidy
```

If this folder ever loses `go.mod`, recreate it from this directory:

```powershell
go mod init download_images
go mod tidy
```

## Run

Open a terminal in `src/collection` and run:

```powershell
go run .
```

You can also run the file directly:

```powershell
go run .\download_images.go
```

## Input And Output

- Input CSV: `../../data/raw/inaturalist/observations-spain-2000-2026.csv`
- Output images: `../../data/images/`

## Typical Workflow

```powershell
Set-Location .\src\collection
go mod tidy
go run .
```
