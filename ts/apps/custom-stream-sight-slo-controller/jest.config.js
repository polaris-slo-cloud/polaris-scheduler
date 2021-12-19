module.exports = {
    displayName: 'custom-stream-sight-slo-controller',
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
    coverageDirectory: '../../coverage/apps/custom-stream-sight-slo-controller',
    reporters: [
        "default",
        [
            "jest-junit", {
                outputDirectory: "./junit-reports",
                outputName: "custom-stream-sight-slo-controller.xml"
            },
        ]
    ],
};
