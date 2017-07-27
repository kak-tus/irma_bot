package Irma::Model::Telegram;

use strict;
use warnings;
use v5.10;
use utf8;

use AnyEvent;
use App::Environ::Config;
use Carp qw(croak);
use Cpanel::JSON::XS;
use Irma::Model::DB;
use Irma::Model::Storage;
use Mojo::Log;
use Mojo::UserAgent;
use Params::Validate qw( validate validate_pos );
use Text::Trim qw(trim);

my %VALIDATION;

BEGIN {
  %VALIDATION = (
    new => {
      config  => 1,
      api_key => 1,
      logger  => 1,
      db      => 1,
      storage => 1,
    },
    process     => { data    => 1 },
    get_message => { chat_id => 1, text => 1, buttons => 0 },
  );
}

use fields keys %{ $VALIDATION{new} };

use Class::XSAccessor { accessors => [ keys %{ $VALIDATION{new} } ] };

my $INSTANCE;

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
    my $api_key = $config->{telegram}{api_key};
    my $logger  = Mojo::Log->new;
    my $db      = Irma::Model::DB->instance;
    my $storage = Irma::Model::Storage->instance;

    $INSTANCE = $class->new(
      config  => $config->{telegram},
      api_key => $api_key,
      logger  => $logger,
      db      => $db,
      storage => $storage,
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

  $self->logger->debug( 'Request: ' . $JSON->encode($form) );

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
    $self->logger->error( 'Request fail: ' . $JSON->encode( $tx->error ) );
    $cb->( undef, 'Request fail' );
    return;
  }

  $cb->( $tx->res->json );

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

  my $msg = $params{data}->{callback_query}{message};

  my ( $user_id, $question_id, $answer_id ) = split '_',
      $params{data}->{callback_query}{data};

  unless ( $user_id == $params{data}->{callback_query}{from}{id} ) {
    $cb->( {} );
    return;
  }

  $self->db->read_group(
    id => $msg->{chat}{id},
    sub {
      $self->_callback_query_group( \%params, $user_id, $question_id,
        $answer_id, @_, $cb );
    }
  );

  return;
}

sub _callback_query_group {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $user_id, $question_id, $answer_id, $group, $err ) = @_;

  if ($err) {
    $cb->( undef, $err );
    return;
  }

  unless ($group) {
    $self->logger->debug('Group not found');
    $cb->( {} );
    return;
  }

  my $correct
      = $group->{questions}[$question_id]{answers}[$answer_id]{correct}
      ? 1
      : 0;

  my $msg     = $params->{data}{callback_query}{message};
  my $chat_id = $msg->{chat}{id};
  my $type    = $msg->{chat}{type};

  if ($correct) {
    $self->logger->debug('Correct answer');
    $self->storage->delete(
      key  => 'kick',
      vals => {
        chat_id => $chat_id,
        user_id => $user_id,
      },
      sub { }
    );
  }
  else {
    $self->_kick_user(
      user_id => $user_id,
      chat_id => $chat_id,
      type    => $type,
      sub { }
    );
  }

  $cb->( {} );

  return;
}

sub _message {
  my __PACKAGE__ $self = shift;

  my $cb = pop;
  croak 'No cb' unless $cb;

  my %params = @_;

  my $msg     = $params{data}->{message};
  my $chat_id = $msg->{chat}{id};
  my $user_id = $msg->{from}{id};

  $self->storage->read(
    key  => 'kick',
    vals => {
      chat_id => $chat_id,
      user_id => $user_id,
    },
    sub { $self->_message_kick_res( \%params, @_, $cb ) }
  );

  return;
}

sub _message_kick_res {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $res, $err ) = @_;

  if ($err) {
    $cb->( undef, $err );
    return;
  }

  my $msg     = $params->{data}->{message};
  my $type    = $msg->{chat}{type};
  my $chat_id = $msg->{chat}{id};
  my $user_id = $msg->{from}{id};

  if ($res) {
    $self->logger->debug('User found in kick pool');

    $cb->( {} );

    $self->_kick_user(
      user_id => $user_id,
      chat_id => $chat_id,
      type    => $type,
      sub { }
    );

    my %form = ( chat_id => $chat_id, message_id => $msg->{message_id} );
    $self->_request( 'deleteMessage', \%form, sub { } );

    return;
  }

  $self->storage->read(
    key  => 'newbie',
    vals => {
      chat_id => $chat_id,
      user_id => $user_id,
    },
    sub { $self->_message_newbie_res( $params, @_, $cb ) }
  );

  return;
}

sub _message_newbie_res {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $res, $err ) = @_;

  if ($err) {
    $cb->( undef, $err );
    return;
  }

  my $msg     = $params->{data}->{message};
  my $type    = $msg->{chat}{type};
  my $chat_id = $msg->{chat}{id};
  my $user_id = $msg->{from}{id};

  my %entities = $self->_search_entities($msg);

  if ($res) {
    $self->logger->debug('Newbie found');
    $self->_message_from_newbie( $params, \%entities, $cb );
    return;
  }

  if ( ( $type eq 'group' || $type eq 'supergroup' )
    && $msg->{new_chat_members} )
  {
    $self->_new_members( $params, $cb );
  }
  elsif ( ( $type eq 'group' || $type eq 'supergroup' )
    && $entities{_bot_name} )
  {
    $self->_message_to_bot( $params, \%entities, $cb );
  }
  elsif ( $type eq 'private' ) {
    my $resp = $self->get_message(
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

sub _message_from_newbie {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $entities ) = @_;

  my $msg     = $params->{data}->{message};
  my $chat_id = $msg->{chat}{id};
  my $user_id = $msg->{from}{id};
  my $type    = $msg->{chat}{type};

  if ( $entities->{url} || $msg->{forward_from} || $msg->{forward_from_chat} )
  {
    $self->logger->debug('Restricted message found');

    $cb->( {} );

    $self->_kick_user(
      user_id => $user_id,
      chat_id => $chat_id,
      type    => $type,
      sub { }
    );

    my %form = ( chat_id => $chat_id, message_id => $msg->{message_id} );
    $self->_request( 'deleteMessage', \%form, sub { } );
  }
  else {
    $self->storage->delete(
      key  => 'newbie',
      vals => {
        chat_id => $chat_id,
        user_id => $user_id,
      },
      sub {
        $cb->( {} );
      }
    );
  }

  return;
}

sub _new_members {
  my __PACKAGE__ $self = shift;
  my $cb               = pop;
  my $params           = shift;

  my $msg = $params->{data}{message};

  if ( $msg->{from}{id} != $msg->{new_chat_members}[0]{id} ) {
    $cb->( {} );
    return;
  }

  $self->db->read_group(
    id => $msg->{chat}{id},
    sub {
      $self->_new_members_res( $params, @_, $cb );
    }
  );

  return;
}

sub _new_members_res {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $group, $err ) = @_;

  if ($err) {
    $cb->( undef, $err );
    return;
  }

  unless ($group) {
    $self->logger->debug('Group not found');
    $cb->( {} );
    return;
  }

  my $res_cb = sub {
    state $cnt= 2;
    $cnt--;
    return if $cnt > 0;
    $cb->( {} );
  };

  if ( $group->{ban_url} ) {
    $self->logger->debug('URL protection');
    $self->_new_members_url( $params, $group, $res_cb );
  }
  else {
    $res_cb->();
  }

  if ( $group->{ban_question} ) {
    $self->logger->debug('Question protection');
    $self->_new_members_question( $params, $group, $res_cb );
  }
  else {
    $res_cb->();
  }

  return;
}

sub _new_members_url {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $group ) = @_;

  my $msg     = $params->{data}{message};
  my $chat_id = $msg->{chat}{id};

  my $res_cb = sub {
    state $cnt = scalar @{ $msg->{new_chat_members} };
    $cnt--;
    return if $cnt > 0;
    $cb->();
  };

  foreach my $user ( @{ $msg->{new_chat_members} } ) {
    $self->storage->create(
      key  => 'newbie',
      ttl  => 86400 * 7,
      vals => {
        chat_id => $chat_id,
        user_id => $user->{id},
      },
      $res_cb
    );
  }

  return;
}

sub _new_members_question {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $group ) = @_;

  $cb->();

  my $msg     = $params->{data}{message};
  my $chat_id = $msg->{chat}{id};
  my $type    = $msg->{chat}{type};

  foreach my $user ( @{ $msg->{new_chat_members} } ) {
    my $question_id = int( rand( @{ $group->{questions} } ) );
    my $question    = $group->{questions}[$question_id];

    my @buttons;
    foreach my $answer ( @{ $question->{answers} } ) {
      my $data = $user->{id} . '_' . $question_id . '_' . scalar(@buttons);
      push @buttons, { text => $answer->{text}, callback_data => $data };
    }

    my $text = '';

    if ( defined $user->{username} ) {
      $text = '@' . $user->{username};
    }
    elsif ( defined $user->{first_name} || defined $user->{last_name} ) {
      $text = '@' . join( ' ', $user->{first_name}, $user->{last_name} );
    }

    $text .= " $group->{greeting}\n\n$question->{text}";

    my $resp = $self->get_message(
      chat_id => $chat_id,
      text    => $text,
      buttons => \@buttons,
    );

    my $method = delete $resp->{method};
    $self->_request( $method, $resp, sub { } );

    $self->storage->create(
      key  => 'kick',
      ttl  => 600,
      vals => {
        chat_id => $chat_id,
        user_id => $user->{id},
      },
      sub { }
    );

    my $timer;
    $timer = AE::timer 60, 0, sub {
      undef $timer;
      $self->_kick_user(
        user_id => $user->{id},
        chat_id => $chat_id,
        type    => $type,
        sub { }
      );
    };
  }

  return;
}

sub _kick_user {
  my __PACKAGE__ $self = shift;
  my $cb               = pop;
  my %params           = @_;

  $self->storage->read(
    key  => 'kick',
    vals => {
      chat_id => $params{chat_id},
      user_id => $params{user_id},
    },
    sub { $self->_kick_user_res( \%params, @_, $cb ) }
  );

  return;
}

sub _kick_user_res {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $res, $err ) = @_;

  if ($err) {
    $cb->( undef, $err );
    return;
  }

  unless ($res) {
    $cb->();
    return;
  }

  my $res_cb = sub {
    state $cnt = 2;
    $cnt--;
    return if $cnt > 0;
    $cb->();
  };

  my $time = time + 86400;

  my %form = (
    user_id    => $params->{user_id},
    chat_id    => $params->{chat_id},
    until_date => $time,
  );

  $self->_request( 'kickChatMember', \%form, $res_cb );

  $self->storage->delete(
    key  => 'kick',
    vals => {
      chat_id => $params->{chat_id},
      user_id => $params->{user_id},
    },
    $res_cb
  );

  return;
}

sub _message_to_bot {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $entities ) = @_;

  my $msg     = $params->{data}{message};
  my $chat_id = $msg->{chat}{id};

  $self->_request(
    'getChatAdministrators',
    { chat_id => $chat_id },
    sub {
      $self->_message_to_bot_permissions( $params, $entities, @_, $cb );
    }
  );

  return;
}

sub _message_to_bot_permissions {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $entities, $res, $err ) = @_;

  if ($err) {
    $cb->( undef, $err );
    return;
  }

  unless ($res) {
    $cb->( {} );
    return;
  }

  my $found_admin;
  my $msg = $params->{data}{message};

  foreach my $user ( @{ $res->{result} } ) {
    next unless $user->{user}{id} == $msg->{from}{id};
    $found_admin = 1;
    last;
  }

  unless ($found_admin) {
    $self->logger->debug('Not admin');
    $cb->( {} );
    return;
  }

  if ( $entities->{_bot_command} ) {
    $self->_message_to_bot_command( $params, $entities, $cb );
  }
  else {
    $self->_message_to_bot_questions( $params, $entities, $cb );
  }

  return;
}

sub _message_to_bot_command {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $entities ) = @_;

  my $cmd = $entities->{_bot_command};
  my $val = $self->config->{texts}{commands}{$cmd};

  unless ($val) {
    $cb->( {} );
    return;
  }

  my $msg     = $params->{data}{message};
  my $chat_id = $msg->{chat}{id};

  $self->db->create_group(
    id            => $chat_id,
    $val->{field} => $val->{value},
    sub {
      $self->_message_to_bot_command_res( $params, $val, @_, $cb );
    }
  );

  return;
}

sub _message_to_bot_command_res {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $val, $res, $err ) = @_;

  if ($err) {
    $cb->( undef, $err );
    return;
  }

  my $msg     = $params->{data}{message};
  my $chat_id = $msg->{chat}{id};

  my $resp = $self->get_message(
    chat_id => $chat_id,
    text    => $val->{text},
  );

  $cb->($resp);

  return;
}

sub _message_to_bot_questions {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $entities ) = @_;

  my $bot_name = '@' . $self->config->{bot_name};
  my $msg      = $params->{data}{message};

  my $text
      = substr( $msg->{text}, length($bot_name) + 1, length( $msg->{text} ) );

  my @lines = split /\n/, $text;

  my $greeting = '';
  my @questions;

  ## TODO refactor to subs
  foreach my $line (@lines) {
    my ( $question_text, $answers_text ) = $line =~ m/^(.+?\?)(.+?)$/;

    unless ( defined $question_text ) {
      $greeting .= "$line\n";
      next;
    }
    unless ( defined $answers_text ) {
      $greeting .= "$line\n";
      next;
    }

    trim $question_text;
    unless ( defined $question_text ) {
      $greeting .= "$line\n";
      next;
    }

    my @answers = trim( split( ';', $answers_text ) );

    if ( length($question_text) && scalar(@answers) ) {
      next if length($question_text) > $self->config->{limits}{question};

      my %question = ( text => $question_text, answers => [] );
      my $found_correct;

      foreach my $answer_text (@answers) {
        next unless length($answer_text);
        next if length($answer_text) > $self->config->{limits}{answer};

        my %answer;
        my $sign = substr( $answer_text, 0, 1 );

        if ( $sign && $sign eq '+' ) {
          $found_correct = 1;
          $answer{correct} = 1;
          $answer{text}
              = trim( substr( $answer_text, 1, length($answer_text) ) );
        }
        else {
          $answer{text} = $answer_text;
        }

        push @{ $question{answers} }, \%answer;
      }

      next unless $found_correct;
      push @questions, \%question;
    }
    else {
      $greeting .= "$line\n";
    }
  }

  my $chat_id = $msg->{chat}{id};

  trim $greeting;

  unless ( length($greeting) > 0
    && length($greeting) < $self->config->{limits}{greeting}
    && scalar(@questions) )
  {
    $self->logger->debug('Greeting not found');
    my $resp = $self->get_message(
      chat_id => $chat_id,
      text    => $self->config->{texts}{fail},
    );
    $cb->($resp);
    return;
  }

  $self->db->create_group(
    id           => $chat_id,
    greeting     => $greeting,
    questions    => \@questions,
    ban_question => 1,
    sub {
      $self->_message_to_bot_set( $params, @_, $cb );
    }
  );

  return;
}

sub _message_to_bot_set {
  my __PACKAGE__ $self = shift;
  my $cb = pop;
  my ( $params, $res, $err ) = @_;

  if ($err) {
    $cb->( undef, $err );
    return;
  }

  my $msg     = $params->{data}{message};
  my $chat_id = $msg->{chat}{id};
  my $resp    = $self->get_message(
    chat_id => $chat_id,
    text    => $self->config->{texts}{set},
  );
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
    foreach my $button ( @{ $params{buttons} } ) {
      push @buttons, [$button];
    }

    my %keyboard = ( inline_keyboard => [@buttons] );
    $response{reply_markup} = $JSON->encode( \%keyboard );
  }

  return \%response;
}

sub _search_entities {
  my __PACKAGE__ $self = shift;
  my $msg = shift;

  my %entities;
  return unless $msg->{entities};

  my $bot_name = '@' . $self->config->{bot_name};

  foreach my $entity ( @{ $msg->{entities} } ) {
    $entity->{_value}
        = substr( $msg->{text}, $entity->{offset}, $entity->{length} );
    $entities{ $entity->{type} } = $entity;

    if ( $bot_name eq $entity->{_value} ) {
      $entities{_bot_name} = $entity;
    }
  }

  foreach my $cmd ( keys %{ $self->config->{texts}{commands} } ) {
    next if index( $msg->{text}, $cmd ) < 0;
    $entities{_bot_command} = $cmd;
  }

  return %entities;
}

1;
