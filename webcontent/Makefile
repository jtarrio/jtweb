all: output/index.html output/comments.js

clean:
	rm -fR output

output:
	mkdir output

output/index.html: html/index.html output
	cp html/index.html output/

output/comments.js: ts/*.ts output
	npx rollup -c

.PHONY: clean