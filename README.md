# GOAPP base

## Prerequisites
* go (golang)
* npm (node)

## Dependencies

* node-sass
* github.com/urfave/negroni
* github.com/unrolled/render

## Instructions

For windows, you may need to get a version of make or use mingw32-make.exe

### Installation
* npm install
* make install

### Compile and Run
* npm run assets
* make

## TODO

* Add Babel
* Add Minification
* Add React
* Add Tests
* Add Configuration System
* Add file watch
* Add Template Layouts
* Add Routing
* Add Database
* Add Authentication
* Add Linting

## Directory Structure

* src - 3rd party libs source
* pkg - packaged 3rd party libs
* bin - binaries
* log - application logs
* node_modules - for node modules used for js/css/image asset pipeline
* public - web root (Do not put things in here, items are copied from theme folder via 'npm run assets'
* theme - theme files
* theme/data - json files for static data per theme
* theme/templates - golang html templates used by render lib
* theme/assets - stores theme css/js/images
