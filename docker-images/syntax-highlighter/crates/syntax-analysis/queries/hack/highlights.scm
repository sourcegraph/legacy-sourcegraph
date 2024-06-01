; Based on https://github.com/nvim-treesitter/nvim-treesitter/blob/master/queries/hack/highlights.scm

((variable) @variable.builtin
  (#eq? @variable.builtin "$this"))

[
  (comment)
  (xhp_comment)
] @comment

((comment) @comment.documentation
  (#match? @comment.documentation "^[*][*][^*].*[*]$"))

"function" @keyword.function

(scope_identifier) @keyword
(visibility_modifier) @keyword

[
  "implements"
  "using"
  "attribute"
  "const"
  "extends"
  "insteadof"
  "trait"
  "throw"
  "yield"
  "is"
  "as"
  "?as"
  "super"
  "where"
  "list"
] @keyword

[
  "class"
  "type"
  "interface"
  "namespace"
  "enum"
] @keyword.type

[
  "async"
  "await"
] @keyword.coroutine

[
  "use"
  "include"
  "include_once"
  "require"
  "require_once"
] @keyword.import

[
  "new"
  "print"
  "echo"
  "newtype"
  "clone"
  "as"
  "concurrent"
  ; "nameof" This is missing in the grammar
] @keyword.operator

"return" @keyword.return

(new_expression
  . (_) @type)

(new_expression
  (qualified_identifier
    (identifier) @type))

(scoped_identifier
  (qualified_identifier
    (identifier) @type))



(xhp_class_attribute
    name: (xhp_identifier) @variable
)

[
  (abstract_modifier)
  (final_modifier)
  (static_modifier)
  (visibility_modifier)
  (xhp_modifier)
  (inout_modifier)
  (reify_modifier)
] @keyword.modifier


[
  "shape"
  "tuple"
  (array_type)
  "bool"
  "float"
  "int"
  "string"
  "arraykey"
  "void"
  "nonnull"
  "mixed"
  "dynamic"
  "noreturn"
  "nothing"
  "num"
] @type.builtin


(shape_type_specifier (open_modifier)) @type.builtin

(field_specifier (optional_modifier) @identifier.operator)

(null) @constant.builtin

[
  (true)
  (false)
] @boolean

[
  (nullable_modifier)
  (soft_modifier)
  (like_modifier)
] @operator


(alias_declaration
  "newtype"
  .
  (_) @type)

(alias_declaration
  "type"
  .
  (_) @type)

(class_declaration
  name: (identifier) @type)

(interface_declaration
  name: (identifier) @type)

(type_parameter
  name: (identifier) @type)

(collection
  (qualified_identifier
    (identifier) @type .))

[
  "@required"
  "@lateinit"
  (attribute_modifier)
] @identifier.attribute

[
  "="
  "??="
  ".="
  "|="
  "^="
  "&="
  "<<="
  ">>="
  "+="
  "-="
  "*="
  "/="
  "%="
  "**="
  "==>"
  "|>"
  "??"
  "||"
  "&&"
  "|"
  "^"
  "&"
  "=="
  "!="
  "==="
  "!=="
  "<"
  ">"
  "<="
  ">="
  "<=>"
  "<<"
  ">>"
  "->"
  "+"
  "-"
  "."
  "*"
  "/"
  "%"
  "**"
  "++"
  "--"
  "!"
  "?:"
  "="
  "??="
  ".="
  "|="
  "^="
  "&="
  "<<="
  ">>="
  "+="
  "-="
  "*="
  "/="
  "%="
  "**="
  "=>"
  ; type modifiers
  "@"
  "?"
  "~"
  "?->"
  "->"
] @operator

(integer) @number

(float) @number.float

(parameter
  (variable) @variable.parameter)

(call_expression
  function: (qualified_identifier
    (identifier) @function.call .))

(call_expression
  function: (scoped_identifier
    (identifier) @function.call .))

(call_expression
  function: (selection_expression
    (qualified_identifier
      (identifier) @function.call .)))

(selection_expression
    (_)
    (xhp_class_identifier) @variable)

(qualified_identifier
  (_) @variable.module
  .
  (_))

; Eplxicitly handle NAN and INF since they are
; not mentioned in grammar
(qualified_identifier
  . (identifier) @variable.builtin  .
  (#eq? @variable.builtin "NAN"))

(qualified_identifier
  . (identifier) @variable.builtin  .
  (#eq? @variable.builtin "INF"))


; Explicitly handle internal amd module since they are not
; not mentioned in grammar
(qualified_identifier
  . (identifier) @keyword  .
  (#eq? @keyword "internal"))
  (qualified_identifier
  . (identifier) @keyword  .
  (#eq? @keyword "module"))

(namespace_declaration
    name:  (qualified_identifier (identifier) @variable.module))

(use_statement
  (qualified_identifier
    (_) @variable.module .)
  (use_clause))

(use_statement
  (use_type
    "namespace")
  (use_clause
    (qualified_identifier
      (identifier) @variable.module .)
    alias: (identifier)? @variable.module))

(use_statement
  (use_type
    "const")
  (use_clause
    (qualified_identifier
      (identifier) @constant .)
    alias: (identifier)? @constant))

(use_statement
  (use_type
    "function")
  (use_clause
    (qualified_identifier
      (identifier) @function .)
    alias: (identifier)? @function))

(use_statement
  (use_type
    "type")
  (use_clause
    (qualified_identifier
      (identifier) @type .)
    alias: (identifier)? @type))

(use_clause
  (use_type
    "namespace")
  (qualified_identifier
    (_) @variable.module .)
  alias: (identifier)? @variable.module)

(use_clause
  (use_type
    "function")
  (qualified_identifier
    (_) @function .)
  alias: (identifier)? @function)

(use_clause
  (use_type
    "const")
  (qualified_identifier
    (_) @constant .)
  alias: (identifier)? @constant)

(use_clause
  (use_type
    "type")
  (qualified_identifier
    (_) @type .)
  alias: (identifier)? @type)

(function_declaration
  name: (identifier) @function)

(method_declaration
  name: (identifier) @function.method)

(type_arguments
  [
    "<"
    ">"
  ] @punctuation.bracket)

[
  "("
  ")"
  "["
  "]"
  "{"
  "}"
  "<<"
  ">>"
] @punctuation.bracket

(xhp_open
  [
    "<"
    ">"
  ] @tag.delimiter)

(xhp_close
  [
    "</"
    ">"
  ] @tag.delimiter)

[
  "."
  ";"
  "::"
  ":"
  ","
] @punctuation.delimiter

(qualified_identifier
  "\\" @punctuation.delimiter)

(ternary_expression
  [
    "?"
    ":"
  ] @keyword.conditional.ternary)

[
  "if"
  "else"
  "elseif"
  "switch"
  "case"
] @keyword.conditional

[
  "try"
  "catch"
  "finally"
] @keyword.exception

[
  "for"
  "while"
  "foreach"
  "do"
  "continue"
  "break"
] @keyword.repeat

[
  (string)
  (xhp_string)
  (heredoc)
] @string

[
  (xhp_open)
  (xhp_close)
] @tag


(type_specifier (qualified_identifier (identifier) @type))
(type_specifier) @type

(variable) @variable
(identifier) @variable
(pipe_variable) @variable

(braced_expression) @none
