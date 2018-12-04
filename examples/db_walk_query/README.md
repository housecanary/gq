An example of loading graphs of objects from a database using a field walker to
determine the graph of objects to load using joins.

The code in the `loader` package is written to be somewhat generic, but is
largely untested, and makes no attempt to support multiple databases.

A more full-fledged implementation of the loader would support multiple
arguments in the directives to allow mapping user supplied arguments, and
controlling how the prefetches happen (i.e. subselects, sequential queries
loaded with id's from the parent level, etc.).

Judicious use of reflection could simplify creating adapters to `DBModel` to
avoid a good deal of boilerplate code.

Implementing all of the above would veer dangerously close to code that should
be in an ORM. A more realistic approach may be to use similar code to integrate
with an ORM and allow the ORM to build and execute the SQL queries.
