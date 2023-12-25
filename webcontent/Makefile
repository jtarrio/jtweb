all: output/index.html output/comments.js

clean:
	rm -fR output

output:
	mkdir output

output/index.html: html/index.html output
	cp html/index.html output/

output/comments.js: ts/comments.ts output
	npx rollup -c

.PHONY: clean