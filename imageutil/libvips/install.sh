#!/bin/sh

BUILD_DIR=build
VIPS_VERSION=8.15.3
VIPS_URL=https://github.com/libvips/libvips/releases/download/v$VIPS_VERSION/vips-$VIPS_VERSION.tar.xz

dnf install -y \
  tar \
  xz \
  cmake \
  meson \
  gcc-c++ \
  glib2-devel \
  expat-devel \
  libjpeg-turbo-devel \
  libpng-devel \
  libtiff-devel \
  giflib-devel \
  libexif-devel \
  libimagequant-devel \
  fftw-devel \
  orc-devel \
  gobject-introspection-devel \
  libwebp-devel \
  libarchive-devel \
  poppler-glib-devel \
  cairo-devel \
  pango-devel \
  librsvg2-devel \
  openjpeg2-devel \
  ImageMagick-devel

curl -sSLf -o - $VIPS_URL | tar -xvJf - \
  && cd vips-$VIPS_VERSION \
  && meson setup $BUILD_DIR --buildtype=release --default-library=static \
  && cd $BUILD_DIR \
  && ninja \
  && ninja test \
  && ninja install
