.PHONY: all login publish pull remote http local clean

all: login publish pull remote http local clean

login:
	@echo ">>> login"
	@remake login -u $(GITHUB_USER) -p $(GITHUB_TOKEN)

publish:
	@echo ">>> publish"
	@remake publish ${GITHUB_USER}/ci.mk:v0.1.0 -f fixtures/ci.mk

pull:
	@echo ">>> pull"
	@remake pull ${GITHUB_USER}/ci.mk -o pulled-ci.mk

remote:
	@echo ">>> run remote"
	@make -f Makefile.remote test

http:
	@echo ">>> run http"
	@make -f Makefile.http test

local:
	@echo ">>> run local"
	@make -f Makefile.local test

clean:
	@echo ">>> clean"
	@rm -rf pulled-ci.mk Makefile.* .remake
