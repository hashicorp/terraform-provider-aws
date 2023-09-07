# Using Regular Expressions

Regular expressions are a powerful tool. However, they are also very expensive in terms of memory. Ensuring correct and useful functionality is the priority but we have a few tips to minimize impact without affecting capabilities.

* **Consider non-regular expressions options.** [`strings.Contains()`](https://pkg.go.dev/strings#Contains), [`strings.Replace()`](https://pkg.go.dev/strings#Replace), and [`strings.ReplaceAll()`](https://pkg.go.dev/strings#ReplaceAll) are dramatically faster and less memory intensive than regular expressions. If one of these will work equally well, use the non-regular expression option.
* **Order character classes consistently.** We use regular expression caching to reduce our memory footprint. This is more effective if character classes are consistently ordered. Since a character class is a set, order does not affect functionality. We have many equivalent regular expressions that only differ by character class order. Below is the order we recommend for consistency:
    1. Numeric range, _i.e._, digits (_e.g._, `0-9`)
    2. Uppercase alphabetic range (_e.g._, `A-Z`, `A-F`)
    3. Lowercase alphabetic range (_e.g._, `a-z`, `a-f`)
    4. Underscore (`_`)
    5. Everything else (except dash, `-`) in ASCII order: `\t\n\r !"#$%&()*+,./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^abcdefghijklmnopqrstuvwxyz{|}~`
    6. _Last_, dash (`-`)
    - Example 1, both equivalent, Wrong order: `[_a-zA-Z0-9-,.]`, Correct: `[0-9A-Za-z_,.-]`
    - Example 2, both equivalent, Wrong order: `[;a-z0-9]`, Correct: `[0-9a-z;]`
* **Inside character classes, avoid unnecessary character escaping.** Go does not complain about extra character escaping but avoid it to improve cache performance. Inside a character class, _most_ characters do not need to be escaped, as Go assumes you mean the literal character.
    * These characters which normally have special meaning in regular expressions, _inside character classes_ do **not** need to be escaped: `$`, `(`, `)`, `*`, `+`, `.`, `?`, `^`, `{`, `|`, `}`.
    * Dash (`-`), when it is last in the character class or otherwise unambiguously not part of a range, does not need to be escaped. If in doubt, place the dash _last_ in the character class (_e.g._, `[a-c-]`) or escape the dash (_e.g._, `\-`).
    * Angle brackets (`[`, `]`) always need to be escaped in a character class.
    * Example 1, both equivalent, Unnecessary escapes: `[\$\(\.\?\|]`, Correct: `[$(.?|]`
    * Example 2, both equivalent, Unnecessary escapes, wrong order: `[a-z\-0-9_A-Z\.]`, Correct: `[0-9A-Za-z_.-]`

<!-- Add links to standard validators to use instead of custom -->
