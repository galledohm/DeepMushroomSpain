# iNaturalist API Scope

This fork targets Spanish mushroom observations between 2000-01-01 and 2026-03-30.

## Exporter Query

The current iNaturalist exporter query used for CSV generation is:

```text
quality_grade=any&identifications=any&iconic_taxa[]=Fungi&place_id=6774&without_taxon_id=54743&rank=species&d1=2000-01-01&d2=2026-03-30
```

This exporter query keeps only species-level fungal observations from Spain and excludes `Lecanoromycetes` so lichen rows do not enter the download pipeline.

## Validated Filters

- place_id: 6774 (Spain)
- taxon_id: 50814 (Agaricomycetes)
- alternate broader taxon_id: 47170 (Fungi)
- excluded exporter taxon_id: 54743 (Lecanoromycetes)
- exporter rank: species
- d1: 2000-01-01
- d2: 2026-03-30
- photos: true
- verifiable: true

## Example Request

```text
https://api.inaturalist.org/v1/observations?place_id=6774&taxon_id=50814&d1=2000-01-01&d2=2026-03-30&photos=true&verifiable=true&per_page=200
```

## Notes

- Responses are paginated JSON.
- The API can be used to build local raw snapshots under data/raw/inaturalist/.
- The exporter and the API are not identical: the exporter query above includes species-rank filtering and excludes `Lecanoromycetes`, while the example API request documents the validated observation-search path for this repo.
- iNaturalist documents a hard limit of 100 requests per minute and asks clients to stay at 60 requests per minute or lower and under 10,000 requests per day.
