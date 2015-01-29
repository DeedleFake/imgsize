Image Size
==========

Image Size is a simple web-app for doing on-the-fly image resizing. It takes a URL, width and height, and a method, and yields a resized version of the image at the URL. The idea is to make it easier to embed images in forums that don't allow image resizing, or other, similar situations.

Requirements
------------

* [resize](https://www.github.com/nfnt/resize): A Go package for resizing images.

Heroku
------

This app is primarily designed for deployment on Heroku, but should be pretty easy to convert to just about any other system. It has no special Heroku-based requirements, and makes no use of any Heroku-specific features.
