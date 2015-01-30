**JX - a tool connecting JSON input to normal text**

JX combines JSON input with a template to produce useful output.

Usage:

        > echo '{"a": "foo"} {"a": "bar"}' | jx 'this is {{a}}'
        this is foo
        this is bar
        >


You can also arrays as json input:

        > echo '["foo"] ["bar"]' | jx 'this is {{1}}'
        this is foo
        this is bar
        >

With arrays the element index is the substitution variable:

        > echo '["foo", "bar"]' | jx 'index 1 is {{1}} and index 2 is {{2}}'
        index 1 is foo and index 2 is bar
        >

Simple JSON types are treated as if they were one element arrays:

        > echo '"foo" 42 true' | jx '{{1}}'
        foo
        42
        true
        >

You can read the template from a file with the -t option:

        > echo 'this is {{a}}' > /tmp/tmpl
        > echo '{"a": "foo"} {"a": "bar"}' | jx -t /tmp/tmpl
        this is foo
        this is bar
        >


You can read the input from a file with the -i option:

        > echo {"a": "foo"} {"a": "bar"}' > /tmp/input
        > jx -i /tmp/input 'this is {{a}}'
        this is foo
        this is bar
        >

You can write output to a designated file with the -o option:


        > echo '{"a": "foo"} {"a": "bar"}' | jx -o /tmp/output 'this is {{a}}'
        > cat /tmp/output
        this is foo
        this is bar
        >


You can use a template to specify the location of the template using the -tx option:

        > echo 'template one is in file {{fn}}' > /tmp/t1
        > echo 'template two is in file {{fn}}' > /tmp/t2
        > echo '{"fn": "t1"} {"fn": "t2"}' | jx -tx /tmp/{{fn}}
        template one is in file t1
        template two is in file t2
        >


Similarly, you can use the -ox option to specify an output filename template:

        > echo '{"fn": "o1"} {"fn": "o2"}' | jx -ox /tmp/{{fn}} 'this is file {{fn}}'
        > cat /tmp/o1
        this is file o1
        > cat /tmp/o2
        this is file o2


Note that by default the -ox option overwrites the previous contents of the file:

        > echo '{"fn": "o1", "a": "first"} {"fn": "o1", "a": "last"}' | jx -ox /tmp/{{fn}} 'this was written {{a}}'
        > cat /tmp/o1
        this was written last


The templates are complete [mustache](https://mustache.github.io/) templates, with the full power that comes
along with them.
