# Metadata Contributions

Use these files for public/community corrections without editing generated metadata.

- `leagues.json`: add missing leagues or override generated league names/logos.
- `teams.json`: add missing teams, aliases, logos, and league mapping.

Example team entry:

```json
{
  "version": 1,
  "last_updated": "2026-04-22",
  "teams": {
    "example_fc": {
      "id": "example_fc",
      "name": "Example FC",
      "logo": "https://example.com/example-fc.png",
      "league": "example_league",
      "aliases": ["Example", "Example Football Club"],
      "source": "community"
    }
  }
}
```

After editing, run:

```powershell
go run main.go scrape
```

The scraper loads generated metadata first, then applies these contribution files as overrides.
