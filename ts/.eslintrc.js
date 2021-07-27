// This config is based on the ESLint configuration used by the Polaris SLO Cloud project (https://github.com/polaris-slo-cloud/polaris), which is licensed under the Apache 2.0 license.

module.exports = {
    "root": true,
    "env": {
        "browser": true,
        "es6": true,
        "node": true
    },
    "plugins": [
        "eslint-plugin-import",
        "eslint-plugin-jsdoc",
        "eslint-plugin-prefer-arrow",
        "@nrwl/nx",
        "@typescript-eslint"
    ],
    "overrides": [
        {
            "files": ["*.ts"],
            "extends": [
                "plugin:@typescript-eslint/recommended",
                "plugin:@typescript-eslint/recommended-requiring-type-checking",
            ],
            "parser": "@typescript-eslint/parser",
            "parserOptions": {
                "project": [
                    "tsconfig.*?.json"
                ],
                "sourceType": "module",
                "ecmaVersion": 2018,
            },
            "rules": {
                "@nrwl/nx/enforce-module-boundaries": [
                    "error",
                    {
                        "enforceBuildableLibDependency": true,
                        "allow": [],
                        "depConstraints": [
                            // Orchestrator-independent core libraries may only depend on each other.
                            { "sourceTag": "scope:core", "onlyDependOnLibsWithTags": [ "scope:core" ] },

                            // Metrics may only depend on core libraries and other metrics.
                            { "sourceTag": "scope:metric", "onlyDependOnLibsWithTags": [ "scope:core", "scope:metric" ] },

                            // SLOs may depend on core, metrics, and other SLOs.
                            // If an SLO only has "scope:slo", it must be orchestrator-independent.
                            // If it should be orchestrator-specific, the tag of the respective orchestrator must be added.
                            // There appears to be a bug in the ESLint plugin that uses the intersection of the allowed libraries of the applied source tags instead of the union.
                            // ToDo: Check if this is fixed in a new version.
                            // { "sourceTag": "scope:slo", "onlyDependOnLibsWithTags": [ "scope:core", "scope:metric", "scope:slo" ] },
                            { "sourceTag": "scope:slo", "onlyDependOnLibsWithTags": [ "*" ] },

                            // Elasticity strategy controllers may only depend on core and other elasticity libraries.
                            // ToDo: Same as for scope:slo.
                            { "sourceTag": "scope:elasticity", "onlyDependOnLibsWithTags": [ "*" ] },

                            // ElasticityStrategies may depend on core, metrics, SLOs, and other strategies.
                            { "sourceTag": "scope:elasticity-strategy", "onlyDependOnLibsWithTags": [ "scope:core", "scope:metric", "scope:slo", "scope:elasticity-strategy" ] },

                            // Kubernetes-specific libraries may only depend on core libraries and other Kubernetes libraries.
                            // A Kubernetes-specific SLO would e.g., have the tags "scope:slo" and "orchestrator:kubernetes"
                            { "sourceTag": "orchestrator:kubernetes", "onlyDependOnLibsWithTags": [ "scope:core", "orchestrator:kubernetes" ] },

                            // CLI apps may only depend on core and the respective orchestrator library that they define using "orchestrator:*".
                            // { "sourceTag": "scope:cli", "onlyDependOnLibsWithTags": [ "scope:core" ] },
                            // There appears to be a bug in the ESLint plugin that uses the intersection of the allowed libraries of the applied source tags instead of the union.
                            // ToDo: Check if this is fixed in a new version.
                            { "sourceTag": "scope:cli", "onlyDependOnLibsWithTags": [ "*" ] },

                            // UI projects may depend on any library project.
                            { "sourceTag": "scope:ui", "onlyDependOnLibsWithTags": [ "*" ] },
                        ]
                    }
                ],
                "@typescript-eslint/array-type": [
                    "error",
                    {
                        "default": "array"
                    }
                ],
                "@typescript-eslint/ban-types": [
                    "error",
                    {
                        "types": {
                            "Object": {
                                "message": "Avoid using the `Object` type. Did you mean `object`?"
                            },
                            "Boolean": {
                                "message": "Avoid using the `Boolean` type. Did you mean `boolean`?"
                            },
                            "Number": {
                                "message": "Avoid using the `Number` type. Did you mean `number`?"
                            },
                            "String": {
                                "message": "Avoid using the `String` type. Did you mean `string`?"
                            },
                            "Symbol": {
                                "message": "Avoid using the `Symbol` type. Did you mean `symbol`?"
                            }
                        }
                    }
                ],
                "@typescript-eslint/consistent-type-definitions": "error",
                "@typescript-eslint/dot-notation": "off",
                "@typescript-eslint/explicit-function-return-type": [
                    "error",
                    {
                        "allowExpressions": true,
                        "allowTypedFunctionExpressions": true,
                        "allowHigherOrderFunctions": true,
                    }
                ],
                "@typescript-eslint/explicit-member-accessibility": [
                    "off",
                    {
                        "accessibility": "explicit"
                    }
                ],
                "@typescript-eslint/explicit-module-boundary-types": [
                    "warn",
                    {
                        "allowArgumentsExplicitlyTypedAsAny": true,
                        "allowDirectConstAssertionInArrowFunctions": true,
                        "allowedNames": [],
                        "allowHigherOrderFunctions": true,
                        "allowTypedFunctionExpressions": true,
                    },
                ],
                // I had to turn this off, because it is currently buggy (https://github.com/typescript-eslint/typescript-eslint/issues/1824 ).
                "@typescript-eslint/indent": "off",
                "@typescript-eslint/member-ordering": [
                    "error",
                    {
                        // The only difference between the default configuration and this one is that
                        // abstract fields and methods should be placed directly after static fields/methods.
                        "default": [
                            // Index signature
                            "signature",

                            // Fields
                            "public-static-field",
                            "protected-static-field",
                            "private-static-field",

                            "public-abstract-field",
                            "protected-abstract-field",
                            "private-abstract-field",

                            // We don't require that decorated fields are placed in a particular order in respect to other instance fields.
                            // "public-decorated-field",
                            // "protected-decorated-field",
                            // "private-decorated-field",

                            "public-instance-field",
                            "protected-instance-field",
                            "private-instance-field",

                            "public-field",
                            "protected-field",
                            "private-field",

                            "static-field",
                            "instance-field",
                            "abstract-field",

                            // "decorated-field",

                            "field",

                            // Constructors
                            "public-constructor",
                            "protected-constructor",
                            "private-constructor",

                            "constructor",

                            // Methods
                            "public-static-method",
                            "protected-static-method",
                            "private-static-method",

                            "public-abstract-method",
                            "protected-abstract-method",
                            "private-abstract-method",

                            "public-decorated-method",
                            "protected-decorated-method",
                            "private-decorated-method",

                            "public-instance-method",
                            "protected-instance-method",
                            "private-instance-method",

                            "public-method",
                            "protected-method",
                            "private-method",

                            "static-method",
                            "instance-method",
                            "abstract-method",

                            "decorated-method",

                            "method"
                        ]
                    }
                ],
                "@typescript-eslint/naming-convention": [
                    "error",
                    {
                        selector: 'default',
                        format: ['camelCase'],
                        leadingUnderscore: 'allow',
                        trailingUnderscore: 'allow',
                    },
                    {
                        selector: 'variable',
                        format: ['camelCase', 'UPPER_CASE'],
                        leadingUnderscore: 'allow',
                        trailingUnderscore: 'allow',
                    },
                    {
                        selector: 'typeLike',
                        format: ['PascalCase'],
                    },
                    {
                        selector: 'enumMember',
                        format: ['PascalCase']
                    },
                ],
                "@typescript-eslint/no-empty-function": "off",
                "@typescript-eslint/no-empty-interface": "off",
                "@typescript-eslint/no-explicit-any": "off",
                "@typescript-eslint/no-inferrable-types": [
                    "error",
                    {
                        "ignoreParameters": true
                    }
                ],
                "@typescript-eslint/no-non-null-assertion": "error",
                "@typescript-eslint/no-parameter-properties": "off",
                "@typescript-eslint/no-unsafe-assignment": "off",
                "@typescript-eslint/no-unsafe-return": "off",
                "@typescript-eslint/no-unused-expressions": "error",
                "@typescript-eslint/no-unused-vars": [
                    "warn",
                    {
                        "args": "none",
                        "vars": "all",
                        "ignoreRestSiblings": true,
                    }
                ],
                "@typescript-eslint/no-use-before-define": "off",
                "@typescript-eslint/no-var-requires": "error",
                "@typescript-eslint/prefer-for-of": "error",
                "@typescript-eslint/prefer-function-type": "error",
                "@typescript-eslint/quotes": [
                    "error",
                    "single"
                ],
                "@typescript-eslint/triple-slash-reference": [
                    "error",
                    {
                        "path": "always",
                        "types": "prefer-import",
                        "lib": "always"
                    }
                ],
                "@typescript-eslint/typedef": [
                    "error",
                    {
                        "arrayDestructuring": false,
                        "arrowParameter": false,
                        "memberVariableDeclaration": false,
                        "objectDestructuring": false,
                        "parameter": true,
                        "propertyDeclaration": true,
                        "variableDeclaration": false,
                        "variableDeclarationIgnoreFunction": false,
                    }
                ],
                "@typescript-eslint/unbound-method": [
                    "error",
                    {
                        "ignoreStatic": true
                    }
                ],
                "@typescript-eslint/unified-signatures": "error",
                "arrow-body-style": "error",
                "arrow-parens": [
                    "off",
                    "always"
                ],
                "comma-dangle": [
                    "error",
                    {
                        "arrays": "always-multiline",
                        "functions": "always-multiline",
                        "imports": "always-multiline",
                        "objects": "always-multiline"
                    }
                ],
                "complexity": "off",
                "constructor-super": "error",
                "eqeqeq": [
                    "error",
                    "smart"
                ],
                "guard-for-in": "error",
                "id-blacklist": [
                    "error",
                    "any",
                    "Number",
                    "number",
                    "String",
                    "string",
                    "Boolean",
                    "boolean",
                    "Undefined",
                ],
                "id-match": "error",
                "import/order": [
                    "error",
                    {
                        alphabetize: {
                            order: 'asc',
                            caseInsensitive: true,
                        }
                    }
                ],
                "jsdoc/check-alignment": "error",
                "jsdoc/newline-after-description": "error",
                "jsdoc/no-types": "error",
                "linebreak-style": [
                    "error",
                    "unix"
                ],
                "max-classes-per-file": "off",
                "max-len": [
                    "warn",
                    {
                        "code": 160
                    }
                ],
                "new-parens": "error",
                "no-bitwise": "error",
                "no-caller": "error",
                "no-cond-assign": "error",
                "no-console": [
                    "error",
                    {
                        "allow": [
                            "log",
                            "warn",
                            "dir",
                            "timeLog",
                            "assert",
                            "clear",
                            "count",
                            "countReset",
                            "group",
                            "groupEnd",
                            "table",
                            "dirxml",
                            "error",
                            "groupCollapsed",
                            "Console",
                            "profile",
                            "profileEnd",
                            "timeStamp",
                            "context"
                        ]
                    }
                ],
                "no-debugger": "error",
                "no-empty": "off",
                "no-eval": "error",
                "no-fallthrough": "error",
                "no-invalid-this": "off",
                "no-multiple-empty-lines": "off",
                "no-new-wrappers": "error",
                "no-restricted-imports": [
                    "error",
                    "rxjs/Rx"
                ],
                "no-shadow": [
                    "error",
                    {
                        "hoist": "all"
                    }
                ],
                "no-throw-literal": "error",
                "no-trailing-spaces": "error",
                "no-undef-init": "error",
                "no-underscore-dangle": "off",
                "no-unsafe-finally": "error",
                "no-unused-labels": "error",
                "no-use-before-define": "off",
                "object-shorthand": "error",
                "one-var": [
                    "error",
                    "never"
                ],
                "prefer-arrow/prefer-arrow-functions": [
                    "error",
                    {
                        "allowStandaloneDeclarations": true,
                    }
                ],
                "quote-props": [
                    "error",
                    "as-needed"
                ],
                "radix": "error",
                "sort-imports": [
                    "error",
                    {
                        // Sorting of multiple imports from one module is provided by sort-imports, sorting the order of the modules is provided by import/order
                        "ignoreDeclarationSort": true,
                    }
                ],
                "spaced-comment": [
                    "error",
                    "always",
                    {
                        "markers": [
                            "/"
                        ]
                    }
                ],
                "use-isnan": "error",
                "valid-typeof": "off",
                // The following are TSLint rule is not (yet) supported by ESLint.
                // It could be run by enabling TSLint integration.
                // "@typescript-eslint/tslint/config": [
                //     "error",
                //     {
                //         "rules": {
                //             "whitespace": [
                //                 true,
                //                 "check-branch",
                //                 "check-decl",
                //                 "check-module",
                //                 "check-operator",
                //                 "check-separator",
                //                 "check-rest-spread",
                //                 "check-type",
                //                 "check-typecast",
                //                 "check-type-operator",
                //                 "check-preblock"
                //             ]
                //         }
                //     }
                // ]
            },
        },
    ]
};
