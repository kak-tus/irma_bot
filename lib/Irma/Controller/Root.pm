package Irma::Controller::Root;

use strict;
use warnings;
use v5.10;
use utf8;

use Mojo::Base 'Mojolicious::Controller';

sub message {
  my $self = shift->openapi->valid_input or return;

  my $v = $self->validation;

  $self->render_later;

  $self->telegram->message(
    data => $v->param('data'),
    sub { $self->_message_res(@_) }
  );

  return;
}

sub _message_res {
  my $self = shift;
  my ( $msg, $err ) = @_;

  if ($err) {
    $self->render( openapi => {}, status => 500 );
    return;
  }

  $self->render( openapi => $msg );

  return;
}
1;
