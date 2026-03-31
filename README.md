# DeepMushroom

DeepMushroom is a fungal classficiation project using [ResNet](https://en.wikipedia.org/wiki/Residual_neural_network)

## Repository Layout

The repository has been re-structured around dataset lifecycle stages and execution concerns:

```text
data/
  raw/inaturalist/         # downloaded observation exports and raw API snapshots
  interim/                 # downloaded images and temporary working assets
  processed/               # model-ready training datasets
docs/
  assets/                  # figures used by the documentation
src/
  collection/              # data acquisition scripts
  training/                # model training entry points
tools/                     # one-off utilities and maintenance scripts
```

## Data Sources

### iNaturalist.org

iNaturalist.org is a citizen science website that allows people to upload images of unknown organisms for identification by other ecology enthusiasts. This fork is now scoped to Spanish mushroom observations between `2000-01-01` and `2026-03-30`.

The historical CSV exports have been moved to `data/raw/inaturalist/`. The image download script now lives in `src/collection/download_images.go`, and the FastAI training entry point now lives in `src/training/train_fastai.py`.

#### iNaturalist Exporter Query

This is the exporter query used to download Spanish fungal observations from `2000-01-01` to `2026-03-30`, restricted to the `species` rank, excluding the lichen class `Lecanoromycetes`:

```text
quality_grade=any&identifications=any&iconic_taxa[]=Fungi&place_id=6774&without_taxon_id=54743&rank=species&d1=2000-01-01&d2=2026-03-30
```

The key constraints in that exporter query are:

- `place_id=6774` limits results to Spain
- `rank=species` excludes genus-level and variety-level observations
- `without_taxon_id=54743` excludes `Lecanoromycetes`, which are the lichen class
- `iconic_taxa[]=Fungi` keeps the export within fungi

#### Using the iNaturalist API instead of the manual export page

Yes. The official observations API supports filtering by place, taxon, date range, and photo availability, so the website export page is not required for this workflow.

Relevant identifiers validated for this fork:

- `place_id=6774` for Spain
- `taxon_id=50814` for Agaricomycetes when you want a mushroom-oriented subset
- `taxon_id=47170` for all fungi if you want the broader fungal kingdom
- `taxon_id=54743` for `Lecanoromycetes`, which can be excluded in exporter-based workflows

Example API query for Spanish mushrooms in the requested date window:

```text
https://api.inaturalist.org/v1/observations?place_id=6774&taxon_id=50814&d1=2015-01-01&d2=2026-01-01&photos=true&verifiable=true&per_page=200
```

Notes:

- The API returns paginated JSON, not a CSV export.
- The public API is rate-limited. iNaturalist documents a hard cap of 100 requests per minute and asks clients to stay at 60 requests per minute or lower and under 10,000 requests per day.
- A broader fungi query with `taxon_id=47170` also works for Spain and the same date range.

#### CSV Fields

Not all of the current CSV columns are required for the image-only workflow.

The current downloader only needs a small subset of fields such as:

- `id`
- `image_url`
- `scientific_name`

The remaining fields are being kept intentionally so the dataset can support future models that may include metadata beyond the image itself, such as location, coordinates, observation date, or other contextual signals.

#### Distribution

![Species distribution](docs/assets/distribution.png)

The data distribution is heavily skewed towards the few most common species. We remove the fungal species with less than 10 images for two reasons:

- If one species has less than 10 identification on iNaturalist.org, it indicates that it is not frequently occuring. Therefore, there is less value in the identification of such species.
- There is not enough data to effectively train the identification model. A species with less than 10 images will hurt the overall accuracy of our model

### MushroomExpert.com

Since the images from MushroomExpert were identified by mycologists, we can use their images as a reliable validator to test the performance of our model.

## Model

Since we are in the very early stage of the experiment we built the model with the [fast.ai](https://www.fast.ai/) library. The model will gradually switch to our own models utilizing [pytorch](https://pytorch.org/) as we progress.

### Metrics

|     Architecture    | Validation Accuracy | Validation Top-5 Accuracy | Test Accurarcy | Test Top-5 Accuracy |
|:-------------------:|:-------------------:|:-------------------------:|:--------------:|:-------------------:|
|       ResNet34      |        70.68        |           86.36           |      31.94     |        48.11        |
|       ResNet50      |        79.67        |           91.76           |      38.77     |        59.14        |
| ResNet50+Focal Loss |        80.24        |           92.32           |      39.48     |        60.45        |

#### Top 10 Most Confused Fungal Species

|        Prediction        |       Ground Truth       |
|:------------------------:|:------------------------:|
|    Fomitopsis mounceae   |    Fomitopsis pinicola   |
|   Pleurotus pulmonarius  |    Pleurotus ostreatus   |
| Dacrymyces chrysospermus |   Tremella mesenterica   |
|   Tremella mesenterica   | Dacrymyces chrysospermus |
|  Laetiporus gilbertsonii |   Laetiporus sulphureus  |
|     Stereum hirsutum     |    Stereum complicatum   |
|     Tremella aurantia    |   Tremella mesenterica   |
|    Ganoderma megaloma    |   Ganoderma applanatum   |
|  Laetiporus cincinnatus  |   Laetiporus sulphureus  |
|   Ganoderma applanatum   |     Ganoderma brownii    |
