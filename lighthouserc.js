module.exports = {
  ci: {
    collect: {
      staticDistDir: './web',
    },
    upload: {
      target: 'temporary-public-storage',
    },
    report: {
      html: {
        directory: './lhci-reports',
      },
    },
  },
};