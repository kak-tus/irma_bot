package Irma::Model::DB;

use strict;
use warnings;
use v5.10;
use utf8;

use App::Environ::Config;
use App::Environ::Mojo::Pg;
use Carp qw(croak);
use Cpanel::JSON::XS;
use Mojo::Log;
use Params::Validate qw(validate);

my %VALIDATION;

BEGIN {
  %VALIDATION = (
    new => {
      logger => 1,
      pg     => 1,
    },
    read_group   => { id => 1, },
    create_group => { id => 1, greeting => 1, questions => 1 },
  );
}

use fields keys %{ $VALIDATION{new} };

use Class::XSAccessor { accessors => [ keys %{ $VALIDATION{new} } ] };

my $INSTANCE;

my $JSON = Cpanel::JSON::XS->new();

App::Environ::Config->register(
  qw(
      irma/migrations.yml
      )
);

sub instance {
  my $class = shift;

  unless ($INSTANCE) {
    my $logger = Mojo::Log->new;
    my $pg     = App::Environ::Mojo::Pg->pg('main');

    $pg->auto_migrate(1);

    $INSTANCE = $class->new( logger => $logger, pg => $pg );
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

sub read_group {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = validate( @_, $VALIDATION{read_group} );

  my $sql = q{
    SELECT greeting, questions
    FROM groups
    WHERE id = ?;
  };

  $self->pg->db->query(
    $sql,
    $params{id},
    sub {
      $self->_read_group_res( @_, $cb );
    }
  );

  return;
}

sub _read_group_res {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $db, $err, $res ) = @_;

  if ($err) {
    $cb->( undef, $err );
    return;
  }

  unless ( $res->rows ) {
    $cb->();
    return;
  }

  my $group = $res->hash;
  $group->{questions} = $JSON->decode( $group->{questions} );

  $cb->($group);

  return;
}

sub create_group {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = validate( @_, $VALIDATION{create_group} );

  my $sql = q{
    INSERT INTO groups
    ( id, greeting, questions ) VALUES ( ?, ?, ? )
    ON CONFLICT (id) DO UPDATE SET
      greeting = EXCLUDED.greeting,
      questions = EXCLUDED.questions;
  };

  my $questions = $JSON->encode( $params{questions} );

  $self->pg->db->query(
    $sql,
    $params{id},
    $params{greeting},
    $questions,
    sub {
      $self->_create_group_res( @_, $cb );
    }
  );

  return;
}

sub _create_group_res {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $db, $err, $res ) = @_;

  if ($err) {
    $cb->( undef, $err );
    return;
  }

  $cb->();

  return;
}

1;
