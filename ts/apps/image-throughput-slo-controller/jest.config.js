module.exports = {
    displayName: 'image-throughput-slo-controller',
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
    coverageDirectory: '../../coverage/apps/image-throughput-slo-controller',
    reporters: [
        "default",
        [
            "jest-junit", {
                outputDirectory: "./junit-reports",
                outputName: "image-throughput-slo-controller.xml"
            },
        ]
    ],
};
