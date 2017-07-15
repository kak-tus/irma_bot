#!/usr/bin/env perl

use strict;
use warnings;
use v5.10;
use utf8;

use Mojo::Base -strict;

$ENV{MOJO_MODE}    = 'development';
$ENV{APPCONF_DIRS} = '/etc:etc:dev/etc';

unshift( @INC, 'lib' );

require Mojolicious::Commands;
Mojolicious::Commands->start_app('Irma::Mojo');
