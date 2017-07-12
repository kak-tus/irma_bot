use strict;
use warnings;
use v5.10;
use utf8;

use ExtUtils::MakeMaker::CPANfile;

WriteMakefile(
  NAME          => 'irma_bot',
  VERSION_FROM  => 'lib/Irma/Mojo.pm',
  ABSTRACT_FROM => 'lib/Irma/Mojo.pm',
  AUTHOR        => 'Andrey Kuzmin',
  EXE_FILES     => ['bin/irma.pl'],
);

package MY;

sub install {
  my $self = shift;

  my $install_sec = $self->SUPER::install(@_);
  $install_sec =~ s/doc_install\n/etc_install data_install\n/;

  $install_sec .= <<MAKE

etc_install :
	\$(NOECHO) \$(MOD_INSTALL) \\
		etc \$(DESTDIR)/etc

data_install :
	\$(NOECHO) \$(MOD_INSTALL) \\
		data \$(DESTDIR)/data

MAKE
      ;

  return $install_sec;
}

1;

