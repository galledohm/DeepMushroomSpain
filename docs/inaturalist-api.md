# iNaturalist API Scope

This fork targets Spanish mushroom observations between 2015-01-01 and 2026-01-01.

## Validated Filters

- place_id: 6774 (Spain)
- taxon_id: 50814 (Agaricomycetes)
- alternate broader taxon_id: 47170 (Fungi)
- d1: 2015-01-01
- d2: 2026-01-01
- photos: true
- verifiable: true

## Example Request

```text
https://api.inaturalist.org/v1/observations?place_id=6774&taxon_id=50814&d1=2015-01-01&d2=2026-01-01&photos=true&verifiable=true&per_page=200
```

## Notes

- Responses are paginated JSON.
- The API can be used to build local raw snapshots under data/raw/inaturalist/.
- iNaturalist documents a hard limit of 100 requests per minute and asks clients to stay at 60 requests per minute or lower and under 10,000 requests per day.
