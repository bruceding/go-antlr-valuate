grammar Govaluate;

expression: primary
            | expression bop='.' ( IDENTIFIER | functionCall)
            | expression bop='[' expression ']'
            | '(' expression ')'
            | functionCall
            | prefix=('+'|'-'|'++'|'--') expression
            | prefix=('~'|'!') expression
            | expression bop=('*'|'/'|'%'|'**'|'^') expression
            | expression bop=('+'|'-') expression
            | expression bop=('<<' |  '>>') expression
            | expression bop=('<=' | '>=' | '>' | '<') expression
            | expression bop=('==' | '!=') expression
            | expression bop='&' expression
            | expression bop='|' expression
            | expression bop='&&' expression
            | expression bop='||' expression
            | expression bop='?' expression ':' expression
            //| <assoc=right> expression
             // bop=('=' | '+=' | '-=' | '*=' | '/=' | '&=' | '|=' | '^=' | '>>=' | '>>>=' | '<<=' | '%=')
              //expression
            | expression bop='in' expression 
           ;

expressionList
    : expression (',' expression)*
    ;

functionCall
    : IDENTIFIER '(' expressionList? ')'
    ;
primary: FLOAT_LITERAL #float
       | STRING_LITERAL #string
       | BOOL_LITERAL #bool
       | array #arrays
       | IDENTIFIER #identifier
       ;

// float 定义
FLOAT_LITERAL: [0-9]+
             | (Digits '.' Digits? | '.' Digits) ExponentPart? [fFdD]?
             |       Digits (ExponentPart [fFdD]? | [fFdD])
             ;
fragment ExponentPart
    : [eE] [+-]? Digits
    ;

fragment Digits
    : [0-9] ([0-9_]* [0-9])?
    ;
// string 定义
STRING_LITERAL:     '\'' (~['\\\r\n] | EscapeSequence)* '\'';
fragment EscapeSequence
    : '\\' [btnfr"'\\]
    | '\\' ([0-3]? [0-7])? [0-7]
    | '\\' 'u'+ HexDigit HexDigit HexDigit HexDigit
    ;

fragment HexDigits
    : HexDigit ((HexDigit | '_')* HexDigit)?
    ;

fragment HexDigit
    : [0-9a-fA-F]
    ;

BOOL_LITERAL: 'true'
            | 'false'
            ;
array : '(' array_value (',' array_value)* ')' ;
array_value : STRING_LITERAL
             | FLOAT_LITERAL
             | BOOL_LITERAL
             ;

// Identifiers
IDENTIFIER:         Letter LetterOrDigit*
          //| '[' Letter LetterOrDigit* ']'
          | '${' Letter LetterOrDigit* '}'
          ;

fragment LetterOrDigit
    : Letter
    | [0-9-_]
    ;

fragment Letter
    : [a-zA-Z$] // these are the "java letters" below 0x7F
    | ~[\u0000-\u007F\uD800-\uDBFF] // covers all characters above 0x7F which are not a surrogate
    | [\uD800-\uDBFF] [\uDC00-\uDFFF] // covers UTF-16 surrogate pairs encodings for U+10000 to U+10FFFF
    ;

WS:     [ \t\r\n]+ -> skip; // 忽略空白字符

