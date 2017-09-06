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
    create => { key => 1, vals => 1, ttl => 0 },
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

sub create {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = validate( @_, $VALIDATION{create} );

  my $key = _key( $params{key}, $params{vals} );
  $params{ttl} //= 600;

  $self->redis->setex(
    $key,
    $params{ttl},
    1,
    sub {
      $self->_res( @_, $cb );
    }
  );

  return;
}

sub _res {
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

sub read {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = validate( @_, $VALIDATION{create} );

  my $key = _key( $params{key}, $params{vals} );

  $self->redis->get(
    $key,
    sub {
      $self->_res( @_, $cb );
    }
  );

  return;
}

sub delete {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = validate( @_, $VALIDATION{create} );

  my $key = _key( $params{key}, $params{vals} );

  $self->redis->del(
    $key,
    sub {
      $self->_res( @_, $cb );
    }
  );

  return;
}

sub _key {
  my ( $key, $vals ) = @_;
  return "irma_${key}_"
      . join( '_', map { $_ . '_' . $vals->{$_} } sort keys %$vals );
}

sub update {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = validate( @_, $VALIDATION{create} );

  my $key = _key( $params{key}, $params{vals} );

  $self->redis->incr(
    $key,
    sub {
      $self->_res( @_, $cb );
    }
  );

  return;
}

1;
