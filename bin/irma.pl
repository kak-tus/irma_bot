#!/usr/bin/env perl

use strict;
use warnings;
use v5.10;
use utf8;

use Mojo::Base -strict;

$ENV{MOJO_MODE}    = 'production';
$ENV{APPCONF_DIRS} = '/etc';

require Mojolicious::Commands;
Mojolicious::Commands->start_app('Irma::Mojo');
