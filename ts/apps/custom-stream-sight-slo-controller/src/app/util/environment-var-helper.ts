export type EnvironmentVariableConverter<T> = (value: string) => T;

export const convertToString: EnvironmentVariableConverter<string> = value => value;

export const convertToNumber: EnvironmentVariableConverter<number> = value => {
    const ret = +value;
    return !Number.isNaN(ret) ? ret : undefined;
};

export function getEnvironmentVariable<T>(name: string, converter: EnvironmentVariableConverter<T>): T;
export function getEnvironmentVariable(name: string): string;
export function getEnvironmentVariable<T = string>(name: string, converter?: EnvironmentVariableConverter<T>): T {
    if (!converter) {
        converter = convertToString as any;
    }
    const value = process.env[name];
    return converter(value);
}
