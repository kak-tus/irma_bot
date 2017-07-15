package Irma::Controller::Root;

use strict;
use warnings;
use v5.10;
use utf8;

use Mojo::Base 'Mojolicious::Controller';

use Mojo::JSON qw(to_json);

sub message {
  my $self = shift->openapi->valid_input or return;

  my $v = $self->validation;

  $self->render_later;

  my $data = $v->param('data');

  $self->logger->debug( 'Message: ' . to_json($data) );

  $self->telegram->process(
    data => $data,
    sub { $self->_message_res(@_) }
  );

  return;
}

sub _message_res {
  my $self = shift;
  my ( $msg, $err ) = @_;

  if ($err) {
    $self->render( json => {}, status => 500 );
    return;
  }

  $self->render( openapi => $msg );

  return;
}

1;
