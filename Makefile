VERSION := 0.0.3

test:
	@ go vet ./...
	@ go run honnef.co/go/tools/cmd/staticcheck@latest ./...
	@ go test -race ./...

precommit: test

release: test
	@ go mod tidy
	@ test -z "`git status --porcelain | grep -vE 'M (History\.md)'`" || (echo "uncommitted changes detected." && false)
	@ test -n "`git status --porcelain | grep -v 'M (History\.md)'`" || (echo "History.md must be uncommited" && false)
	@ git commit -am "Release v$(VERSION)"
	@ git tag "v$(VERSION)"
	@ git push origin main "v$(VERSION)"
	@ go run github.com/cli/cli/v2/cmd/gh@5023b61 release create --notes-file=History.md "v$(VERSION)"