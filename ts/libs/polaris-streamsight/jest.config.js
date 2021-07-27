module.exports = {
    displayName: 'polaris-streamsight',
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
    coverageDirectory: '../../coverage/libs/polaris-streamsight',
    reporters: [
        "default",
        [
            "jest-junit", {
                outputDirectory: "./junit-reports",
                outputName: "polaris-streamsight.xml"
            },
        ]
    ],
};
