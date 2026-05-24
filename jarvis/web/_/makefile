.PHONY: dist

dist:
	rm -rf dist
	mkdir -p dist/
	cp -RHL lib dist/
	rm -f jarvis.zip
	cd dist && zip -r ../jarvis.zip lib -x '*/.DS_Store'
