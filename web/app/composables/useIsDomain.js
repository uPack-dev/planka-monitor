export default () => {
  const { host } = useRequestURL();

  return host;
};
