module.exports = {
  devServer: {
    port: 8081,
    headers: { "Cache-Control": "no-cache, no-store, must-revalidate", "Access-Control-Allow-Origin": "*", "Access-Control-Allow-Headers": "*", "Access-Control-Allow-Methods": "*"},
  },
};
