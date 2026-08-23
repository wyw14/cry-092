$ErrorActionPreference = 'Stop'
$changed = gofmt -l .
if ($changed) { throw "gofmt required: $changed" }
go build ./...
go test ./...
go test -race ./...
go vet ./...
Push-Location web
try {
  npm ci
  npm run test:unit -- --run
  npm run check:types
  npm run desk:bundle
} finally {
  Pop-Location
}
