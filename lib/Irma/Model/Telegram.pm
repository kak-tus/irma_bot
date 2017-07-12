package Irma::Model::Telegram;

use strict;
use warnings;
use v5.10;
use utf8;

use App::Environ::Config;
use Carp qw(croak);
use Mojo::UserAgent;
use Params::Validate qw( validate validate_pos );
use Mojo::Log;

my %VALIDATION;

BEGIN {
  %VALIDATION = (
    new => {
      config  => 1,
      api_key => 1,
      logger  => 1,
    },
    message => { data => 1 },
  );
}

use fields keys %{ $VALIDATION{new} };

use Class::XSAccessor { accessors => [ keys %{ $VALIDATION{new} } ] };

my $INSTANCE;

our $MODE = 'production';

App::Environ::Config->register(
  qw(
      irma/irma.yml
      )
);

sub instance {
  my $class = shift;

  unless ($INSTANCE) {
    my $config  = App::Environ::Config->instance->{irma};
    my $api_key = $config->{telegram}{api_keys}{$MODE};

    my $logger = Mojo::Log->new;

    $logger->debug("Start in $MODE mode");

    $INSTANCE = $class->new(
      config  => $config->{telegram},
      api_key => $api_key,
      logger  => $logger,
    );
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

sub init {
  my __PACKAGE__ $self = shift;

  my ($url) = validate_pos( @_, 1 );

  my $form = { url => $url };

  $self->_request( 'setWebhook', $form, sub { } );

  return;
}

sub _request {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $action, $form ) = @_;

  my $api_key = $self->api_key;
  my $uri     = "https://api.telegram.org/bot$api_key/$action";

  my $ua = Mojo::UserAgent->new;

  $ua->post(
    $uri,
    form => $form,
    sub {
      $self->_request_results( $ua, @_, $cb );
    }
  );

  return;
}

sub _request_results {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $ua, undef, $tx ) = @_;

  $self->logger->debug( 'Response: ', $tx->res->body );

  if ( $tx->error ) {
    $self->logger->error( 'Request fail ', $tx->error->{message} );
    $cb->( undef, 'Request fail' );
    return;
  }

  $cb->();

  return;
}

sub message {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = validate( @_, $VALIDATION{message} );

  my $chat_id = $params{data}->{message}{chat}{id};
  $cb->( {} );

  return;
}

sub _message_res {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $user_id, $err ) = @_;

  if ($err) {
    $cb->( undef, $err );
    return;
  }

  unless ($user_id) {
    $self->_check_token( $params, $cb );
    return;
  }

  my $resp = $self->get_message(
    chat_id => $params->{data}{message}{chat}{id},
    text    => $self->config->{texts}{already},
  );
  $cb->($resp);

  return;
}

1;
