package Irma::Model::Storage;

use strict;
use warnings;
use v5.10;
use utf8;

use AnyEvent::RipeRedis;
use App::Environ::Config;
use Carp qw(croak);
use Mojo::Log;
use Params::Validate qw(validate);

my %VALIDATION;

BEGIN {
  %VALIDATION = (
    new => {
      logger => 1,
      redis  => 1,
    },
    kick => { user_id => 1, chat_id => 1 },
  );
}

use fields keys %{ $VALIDATION{new} };

use Class::XSAccessor { accessors => [ keys %{ $VALIDATION{new} } ] };

my $INSTANCE;

my $JSON = Cpanel::JSON::XS->new()->allow_blessed();

App::Environ::Config->register(
  qw(
      redis.yml
      )
);

sub instance {
  my $class = shift;

  unless ($INSTANCE) {
    my $logger = Mojo::Log->new;
    my $config = App::Environ::Config->instance->{redis}{connectors}{local};
    my $redis  = AnyEvent::RipeRedis->new(%$config);

    $INSTANCE = $class->new( logger => $logger, redis => $redis );
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

sub create_kick {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = validate( @_, $VALIDATION{kick} );

  my $key = "irma_kick_$params{chat_id}_$params{user_id}";

  $self->redis->setex(
    $key, 600, 1,
    sub {
      $self->_kick_res( @_, $cb );
    }
  );

  return;
}

sub _kick_res {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $res, $err ) = @_;

  if ($err) {
    $self->logger->error( $JSON->encode($err) );
    $cb->( undef, $err );
    return;
  }

  $cb->($res);

  return;
}

sub read_kick {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = validate( @_, $VALIDATION{kick} );

  my $key = "irma_kick_$params{chat_id}_$params{user_id}";

  $self->redis->exists(
    $key,
    sub {
      $self->_kick_res( @_, $cb );
    }
  );

  return;
}

sub delete_kick {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = validate( @_, $VALIDATION{kick} );

  my $key = "irma_kick_$params{chat_id}_$params{user_id}";

  $self->redis->del(
    $key,
    sub {
      $self->_kick_res( @_, $cb );
    }
  );

  return;
}

1;
