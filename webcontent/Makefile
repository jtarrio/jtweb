all: generated/comments.js

clean:
	rm -fR generated

generated:
	mkdir generated

generated/comments.js: ts/*.ts generated
	npx rollup -c

.PHONY: clean