#!/usr/bin/perl
# Read the skill folders under skills/ of the subject (the working directory)
# and print one check line per rule. A rule holds when one paragraph, list
# item, heading or table row of the read files matches every pattern of the
# rule, case-insensitive. Every line prints a value: with no skill, each rule
# prints "no".
#
#   --scope all    read every .md file under skills/, at any depth
#   --skill NAME   the skill folder skills/NAME
#   --scope part   read only the .md files of skills/NAME, other than SKILL.md,
#                  whose file name SKILL.md names (the run part); say whether
#                  the part shows a hyp command in code, and list every other
#                  skill folder that repeats a paragraph of the part
#   --rules FILE   one rule per line: name, then patterns, split by tabs
#   --forbidden FILE  one reference per line: name, then a pattern (case-sensitive)
#                  that a skill used in another repository must not hold
#   --hyp BIN      the hyp binary built from the subject: every hyp command the
#                  skill shows in code must be one its help lists
#   --probe        check that one executable file of the skill works as the
#                  generic Probe
use strict;
use warnings;
use File::Find;
use File::Basename qw(basename dirname);
use File::Spec;
use Cwd qw(abs_path);
use File::Temp qw(tempdir);

my %o = (scope => 'all');
while (@ARGV) {
    my $k = shift @ARGV;
    if    ($k eq '--scope')     { $o{scope} = shift @ARGV }
    elsif ($k eq '--rules')     { $o{rules} = shift @ARGV }
    elsif ($k eq '--forbidden') { $o{forbidden} = shift @ARGV }
    elsif ($k eq '--hyp')       { $o{hyp} = shift @ARGV }
    elsif ($k eq '--probe')     { $o{probe} = 1 }
    elsif ($k eq '--skill')     { $o{skill} = shift @ARGV }
    else  { die "unknown flag $k\n" }
}

sub slurp { my ($f) = @_; open my $h, '<', $f or return ''; local $/; my $t = <$h>; return defined $t ? $t : '' }

# Skill folders: folders right under skills/ that hold a SKILL.md.
my @folders = grep { -d $_ && -f "$_/SKILL.md" } sort glob('skills/*');
my $skill = "skills/$o{skill}";
my $has = -f "$skill/SKILL.md";
print "skill folder $o{skill}: ", ($has ? 'yes' : 'no'), "\n" if $o{scope} eq 'all';

my @all;
if (-d 'skills') {
    find({ no_chdir => 1, wanted => sub { push @all, $File::Find::name if -f $_ && /\.md\z/ } }, 'skills');
}
@all = sort @all;

my @files;
if ($o{scope} eq 'part') {
    my @named;
    if ($has) {
        my $text = slurp("$skill/SKILL.md");
        for my $f (grep { index($_, "$skill/") == 0 && basename($_) ne 'SKILL.md' } @all) {
            push @named, $f if index($text, basename($f)) >= 0;
        }
    }
    print 'a file beside SKILL.md that SKILL.md names: ', (@named ? 'yes' : 'no'), "\n";
    @files = @named;
} else {
    @files = @all;
}

# Units: a paragraph, a list item, a heading or a table row, with its lines
# joined and its spaces collapsed.
sub units {
my @units;
for my $f (@_) {
    my @cur;
    my $flush = sub { if (@cur) { my $u = join(' ', @cur); $u =~ s/\s+/ /g; push @units, $u; @cur = () } };
    for my $line (split /\n/, slurp($f)) {
        if ($line =~ /^\s*$/) { $flush->(); next }
        if ($line =~ /^\s*([-*+]|\d+[.)])\s/ || $line =~ /^#/ || $line =~ /^\s*\|/) { $flush->() }
        push @cur, $line;
        $flush->() if $line =~ /^#/ || $line =~ /^\s*\|/;
    }
    $flush->();
}
return @units;
}
my @units = units(@files);

# rules: the names of the rules that hold in the given units
my @rules = map { [split /\t/] } grep { !/^\s*$/ } split /\n/, ($o{rules} ? slurp($o{rules}) : '');
sub holds {
    my @u = @_;
    my %h;
    for my $r (@rules) {
        my ($name, @pats) = @$r;
        UNIT: for my $x (@u) {
            for my $p (@pats) { next UNIT unless $x =~ /$p/i }
            $h{$name} = 1;
            last;
        }
    }
    return %h;
}

my %hit = holds(@units);
print "rule $_->[0]: ", ($hit{$_->[0]} ? 'yes' : 'no'), "\n" for @rules;

# hyp commands shown in code: inline code spans and fenced blocks of .md files
sub commands {
    my %named;
    for my $f (@_) {
        my $t = slurp($f);
        my @code;
        push @code, $1 while $t =~ /```[a-z]*\n(.*?)```/gs;
        push @code, $1 while $t =~ /`([^`\n]+)`/g;
        for my $c (@code) {
            while ($c =~ /(?:^|[\s"\$(])(?:hyp|"\$HYP"|\$HYP)\s+([a-z][a-z-]*)/g) { $named{$1} = 1 }
        }
    }
    return %named;
}

if ($o{scope} eq 'part') {
    my %c = commands(@files);
    print 'the run part shows a hyp command in code: ', (%c ? 'yes' : 'no'), "\n";
    # A paragraph of the part, of at least 40 characters, that another skill
    # folder repeats word for word, ignoring case.
    my %part = map { lc($_) => 1 } grep { length($_) >= 40 } @units;
    my @other;
    for my $d (grep { $_ ne $skill } sort grep { -d $_ } glob('skills/*')) {
        my @u = units(grep { index($_, "$d/") == 0 } @all);
        push @other, $d if grep { $part{lc $_} } @u;
    }
    print 'run part text repeated in another skill folder: ', (@other ? join(', ', @other) : 'none'), "\n";
}

my $text = join("\n", map { slurp($_) } @all);

if ($o{forbidden}) {
    for my $r (split /\n/, slurp($o{forbidden})) {
        next if $r =~ /^\s*$/;
        my ($name, $p) = split /\t/, $r, 2;
        print "cites $name: ", ($text =~ /$p/ ? 'yes' : 'no'), "\n";
    }
    # Relative links must resolve to a file inside the skill's own folder.
    my @bad;
    for my $f (@all) {
        my $t = slurp($f);
        while ($t =~ /\]\(([^)\s]+)\)/g) {
            my $l = $1;
            next if $l =~ /^[a-z][a-z0-9+.-]*:/i || $l =~ /^#/;
            $l =~ s/#.*//;
            my $top = $f; $top =~ s{^(skills/[^/]+)/.*}{$1};
            my $p = abs_path(File::Spec->catfile(dirname($f), $l));
            my $root = abs_path($top);
            push @bad, "$f: $l" unless defined $p && -e $p && index($p, "$root/") == 0;
        }
    }
    print 'links that leave the skill folder or do not resolve: ', (@bad ? join(', ', @bad) : 'none'), "\n";
}

if ($o{hyp}) {
    my %known;
    my $in = 0;
    for (split /\n/, `"$o{hyp}" help 2>/dev/null`) {
        if (/^commands:/) { $in = 1; next }
        last if $in && /^\S/;
        $known{$1} = 1 if $in && /^  ([a-z][a-z-]*)/;
    }
    my %named = commands(@all);
    my @unknown = grep { !$known{$_} } sort keys %named;
    print 'hyp commands the skill shows in code: ', (%named ? 'some' : 'none'), "\n";
    print 'hyp commands that hyp help does not list: ', (@unknown ? join(', ', @unknown) : 'none'), "\n";
}

if ($o{probe}) {
    my @exe;
    for my $d (@folders) {
        find({ no_chdir => 1, wanted => sub { push @exe, $File::Find::name if -f $_ && -x $_ } }, $d);
    }
    my $works = 0;
    for my $e (sort @exe) {
        my $dir = tempdir(CLEANUP => 1);
        my $abs = abs_path($e);
        local $ENV{HYP_RUN} = 'printf from-run; echo to-stderr >&2; exit 3';
        my $out = `cd "$dir" && "$abs" 2>/dev/null`;
        if ($out eq "from-run\nexit 3\n") { $works = 1; last }
    }
    print 'a file of the skill works as the generic Probe: ', ($works ? 'yes' : 'no'), "\n";
}
