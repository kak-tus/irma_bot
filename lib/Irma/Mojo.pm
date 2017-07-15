package Irma::Mojo;

our $VERSION = 0.1;

use strict;
use warnings;
use v5.10;
use utf8;

use Mojo::Base 'Mojolicious';

use App::Environ;
use App::Environ::Config;
use Irma::Model::Telegram;

my $LOADED;

register();

sub register {
  App::Environ::Config->register(
    qw(
        sys.yml
        irma/irma.yml
        irma/conf.d/*.yml
        )
  );

  return;
}

sub startup {
  my $self = shift;

  $self->_init_config();
  $self->_init_plugins();
  $self->_init_routes();
  $self->_init_models();
  $self->_init_telegram();

  return;
}

sub _init_config {
  my $self = shift;

  App::Environ->send_event('initialize');
  $LOADED = 1;

  my $config = App::Environ::Config->instance;
  $self->config($config);

  $self->mode( $ENV{MOJO_MODE} // 'production' );

  $self->secrets( $self->config->{irma}{secrets} );

  return;
}

sub _init_plugins {
  my $self = shift;

  if ( $self->mode eq 'production' ) {
    $self->plugin(
      SetUserGroup => {
        user  => $self->config->{irma}{user},
        group => $self->config->{irma}{group}
      }
    );
  }

  return;
}

sub _init_routes {
  my $self = shift;

  my $route = $self->routes();
  $route->namespaces( ['Irma::Controller'] );

  my $api = $self->config->{irma}{api};

  $self->plugin(
    OpenAPI => {
      route => $route,
      url   => $api,
    }
  );

  return;
}

sub _init_models {
  my $self = shift;

  $Irma::Model::Telegram::MODE = $self->mode;

  $self->helper(
    telegram => sub {
      return Irma::Model::Telegram->instance;
    }
  );

  return;
}

sub _init_telegram {
  my $self = shift;

  Mojo::IOLoop->next_tick(
    sub {
      return unless $LOADED;

      my $notify_key
          = $self->config->{irma}{telegram}{notify_keys}{ $self->mode };

      my $url
          = $self->config->{irma}{telegram}{notify_url}
          . $self->url_for( 'telegram_message', notify_key => $notify_key )
          ->to_abs->to_string;

      $self->telegram->init($url);
    }
  );

  return;
}

sub END {
  return unless $LOADED;

  undef $LOADED;

  App::Environ->send_event('finalize:r');

  return;
}

1;

=encoding utf-8

=head1 NAME

Irma::Mojo - @irma_bot, Irma AntiSpam Bot for Telegram groups

=head1 AUTHOR

Andrey Kuzmin

=cut
