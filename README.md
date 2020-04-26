# semvertools
This is a collection of tools for working with semantic version numbers. For
full help on using the commands pass each one a `-help` parameter and you
will see a, hopefully useful, help message giving the parameters that the
command can take and a brief description of the command and what it does.

They all make use of the `github.com/nickwells/semver.mod/semver` and
`github.com/nickwells/semverparam.mod/v4/semverparam` packages and serve as
examples of how to use those packages.


## semvercheck
This will check that a set of semantic version numbers are all well
formed. It can also check that any pre-release IDs or build IDs conform to
user-specified rules. It can also check that the set of semvers are correctly
ordered and that they have no gaps.

## semverincr
This will increment the specified part of the semver and set the other parts
correctly. It can also be used to increment the pre-release ID parts if
appropriate. This could be of use as part of a script.

## semversort
This will correctly sort a set of semvers. This is trickier that it might
appear as there are some slightly complex rules around the ordering of
pre-release IDs.
