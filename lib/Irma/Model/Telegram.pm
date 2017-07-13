package Irma::Model::Telegram;

use strict;
use warnings;
use v5.10;
use utf8;

use App::Environ::Config;
use Carp qw(croak);
use Cpanel::JSON::XS;
use Mojo::Log;
use Mojo::UserAgent;
use Params::Validate qw( validate validate_pos );

# use Irma::Model::DB;

my %VALIDATION;

BEGIN {
  %VALIDATION = (
    new => {
      config  => 1,
      api_key => 1,
      logger  => 1,
    },
    process     => { data    => 1 },
    get_message => { chat_id => 1, text => 1, buttons => 0 },
  );
}

use fields keys %{ $VALIDATION{new} };

use Class::XSAccessor { accessors => [ keys %{ $VALIDATION{new} } ] };

my $INSTANCE;

our $MODE = 'production';

my $JSON = Cpanel::JSON::XS->new;

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

  $self->logger->debug( 'Response: ' . $tx->res->body );

  if ( $tx->error ) {
    $self->logger->error( 'Request fail ', $tx->error->{message} );
    $cb->( undef, 'Request fail' );
    return;
  }

  $cb->();

  return;
}

sub process {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = validate( @_, $VALIDATION{process} );

  if ( $params{data}->{message} ) {
    $self->_message( %params, $cb );
  }
  elsif ( $params{data}->{callback_query} ) {
    $self->_callback_query( %params, $cb );
  }
  else {
    $cb->( {} );
  }

  return;
}

sub _callback_query {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = @_;

  my $resp = $self->get_message(
    chat_id => $params{data}->{callback_query}{message}{chat}{id},
    text    => 'ok',
  );
  $cb->($resp);

  return;
}

sub _message {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = @_;

  my $msg  = $params{data}->{message};
  my $type = $msg->{chat}{type};

  if ( $type eq 'group' && $msg->{new_chat_members} ) {
    $self->_new_members( \%params, $cb );
  }
  elsif ( $type eq 'group' && $msg->{entities} ) {
    $self->_message_to_bot( \%params, $cb );
  }
  elsif ( $type eq 'private' ) {
    my $chat_id = $msg->{chat}{id};
    my $resp    = $self->get_message(
      chat_id => $chat_id,
      text    => $self->config->{texts}{usage},
    );
    $cb->($resp);
  }
  else {
    $cb->( {} );
  }

  return;
}

sub _new_members {
  my __PACKAGE__ $self = shift;
  my $cb               = pop;
  my $params           = shift;

  my $msg = $params->{data}{message};

  # if ( $msg->{from}{id}
  #   != $msg->{new_chat_members}[0]{id} )
  # {
  #   $cb->( {} );
  #   return;
  # }

  $cb->( {} );

  my $chat_id = $msg->{chat}{id};

  foreach my $user ( @{ $msg->{new_chat_members} } ) {
    my $resp = $self->get_message(
      chat_id => $chat_id,
      text    => '@' . "$user->{username} тест",
      buttons => [ "1", "23" ],
    );

    my $method = delete $resp->{method};
    $self->_request( $method, $resp, sub { } );
  }

  return;
}

sub _message_to_bot {
  my __PACKAGE__ $self = shift;
  my $cb               = pop;
  my $params           = shift;

  my $resp = {};

  my $msg = $params->{data}{message};

  foreach my $entity ( @{ $msg->{entities} } ) {
    next unless $entity->{type} eq 'mention';

    my $username
        = substr( $msg->{text}, $entity->{offset}, $entity->{length} );
    next unless $username && $username eq '@' . $self->config->{bot_name};

    my $chat_id = $msg->{chat}{id};
    $resp = $self->get_message(
      chat_id => $chat_id,
      text    => 'set',
    );
    last;
  }

  $cb->($resp);

  return;
}

sub get_message {
  my __PACKAGE__ $self = shift;

  my %params = validate( @_, $VALIDATION{get_message} );

  my %response = (
    method                   => 'sendMessage',
    chat_id                  => $params{chat_id},
    text                     => $params{text},
    disable_web_page_preview => Cpanel::JSON::XS::true,
  );

  if ( $params{buttons} && scalar( @{ $params{buttons} } ) ) {
    my @buttons;
    for ( my $i = 0; $i < scalar( @{ $params{buttons} } ); $i++ ) {
      my $button = $params{buttons}->[$i];
      push @buttons, [ { text => $button, callback_data => "$i" } ];
    }

    my %keyboard = ( inline_keyboard => [@buttons] );
    $response{reply_markup} = $JSON->encode( \%keyboard );
  }

  return \%response;
}

1;
