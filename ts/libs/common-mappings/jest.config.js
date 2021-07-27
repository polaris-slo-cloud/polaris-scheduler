module.exports = {
    displayName: 'common-mappings',
    preset: '../../jest.preset.js',
    globals: {
        'ts-jest': {
            tsconfig: '<rootDir>/tsconfig.spec.json',
        },
    },
    testEnvironment: 'node',
    transform: {
        '^.+\\.[tj]sx?$': 'ts-jest',
    },
    moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx'],
    coverageDirectory: '../../coverage/libs/common-mappings',
    reporters: [
        "default",
        [
            "jest-junit", {
                outputDirectory: "./junit-reports",
                outputName: "common-mappings.xml"
            },
        ]
    ],
};
