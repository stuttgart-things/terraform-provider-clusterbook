# CI/CD

## GitHub Actions Workflows

| Workflow | Trigger | Description |
|----------|---------|-------------|
| `release` | Push to main | semantic-release: changelog, GitHub release with GoReleaser binaries |
| `pages` | After release | Deploy TechDocs to GitHub Pages |

## Release Process

Releases are fully automated via [semantic-release](https://semantic-release.gitbook.io/):

- `fix:` commits trigger a **patch** bump
- `feat:` commits trigger a **minor** bump
- GoReleaser builds binaries for linux/darwin/windows (amd64/arm64)

## Workflow Chain

```
push to main → release → pages
                  │         │
            semantic     techdocs
            release      deploy
            + goreleaser
            binaries
```

## Links

- [Releases](https://github.com/stuttgart-things/terraform-provider-clusterbook/releases)
- [Pages](https://stuttgart-things.github.io/terraform-provider-clusterbook)
