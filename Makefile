default: test

test: vet simple staticcheck $(GOPATH)/bin/ginkgo
	go test ./pkg

integration-test: vet simple staticcheck $(GOPATH)/bin/ginkgo
	go test ./internal

vet:
	@go vet ./pkg

simple: $(GOPATH)/bin/gosimple
	@$(GOPATH)/bin/gosimple ./pkg

errcheck: $(GOPATH)/bin/errcheck
	@$(GOPATH)/bin/errcheck ./pkg

staticcheck: $(GOPATH)/bin/staticcheck
	@$(GOPATH)/bin/staticcheck ./pkg

$(GOPATH)/bin/gosimple:
	go get honnef.co/go/tools/cmd/gosimple

$(GOPATH)/bin/staticcheck:
	go get honnef.co/go/tools/cmd/staticcheck

$(GOPATH)/bin/errcheck:
	go get github.com/kisielk/errcheck

$(GOPATH)/bin/ginkgo:
	go get github.com/onsi/ginkgo/ginkgo

update-defs: definitions/
	git submodule update --init --remote --recursive

definitions: point-to-defs-master

point-to-defs-branch:
	git submodule add -b $(BRANCH) git@github.com:$(USER)/definitions.git definitions/

point-to-defs-master:
	git submodule add https://github.com/holidays/definitions definitions/

clean-defs:
	git rm -f definitions
	rm -rf .git/modules/definitions
	git config -f .git/config --remove-section submodule.definitions 2> /dev/null

.PHONY: test update-defs clean-defs point-to-defs-master point-to-defs-branch definitions

