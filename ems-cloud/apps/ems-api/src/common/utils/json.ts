export function enableBigIntJson() {
  Object.defineProperty(BigInt.prototype, 'toJSON', {
    configurable: true,
    value() { return this.toString(); },
  });
}
