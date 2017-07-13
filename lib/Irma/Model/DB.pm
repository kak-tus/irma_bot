package Irma::Model::DB;

use strict;
use warnings;
use v5.10;
use utf8;

use App::Environ::Mojo::Pg;
use Carp qw(croak);
use Mojo::Log;
use Params::Validate qw( validate validate_pos );

my %VALIDATION;

BEGIN {
  %VALIDATION = (
    new => {
      logger => 1,
      pg     => 1,
    },
  );
}

use fields keys %{ $VALIDATION{new} };

use Class::XSAccessor { accessors => [ keys %{ $VALIDATION{new} } ] };

my $INSTANCE;

sub instance {
  my $class = shift;

  unless ($INSTANCE) {
    my $logger = Mojo::Log->new;
    my $pg     = App::Environ::Mojo::Pg->pg('main');

    $INSTANCE = $class->new( logger => $logger, );
  }

  return $INSTANCE;
}

sub new {
  my $class = shift;

  my %params = validate( @_, $VALIDATION{new} );

  my __PACKAGE__ $self = fields::new($class);

  foreach ( keys %{ $VALIDATION{new} } ) {
    $self->{$_} = $params{$_};
  }

  return $self;
}

1;
