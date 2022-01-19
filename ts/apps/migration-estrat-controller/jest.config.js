module.exports = {
    displayName: 'migration-estrat-controller',
    preset: '../../jest.preset.js',
    globals: {
        'ts-jest': {
            tsconfig: '<rootDir>/tsconfig.spec.json',
        },
    },
    testEnvironment: 'node',
    transform: {
        '^.+\\.[tj]s$': 'ts-jest',
    },
    moduleFileExtensions: ['ts', 'js', 'html'],
    coverageDirectory: '../../coverage/apps/migration-estrat-controller',
    reporters: [
        "default",
        [
            "jest-junit", {
                outputDirectory: "./junit-reports",
                outputName: "migration-estrat-controller.xml"
            },
        ]
    ],
};
