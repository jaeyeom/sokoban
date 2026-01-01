.PHONY: clean

node_modules:
	npm init -y
	npm install --save-dev webpack webpack-cli babel-loader @babel/core @babel/preset-env
	npm install phaser whatwg-fetch buffer process stream-browserify

dist: node_modules
	npx webpack --config webpack.config.js

clean:
	rm -rf node_modules dist package.json package-lock.json
