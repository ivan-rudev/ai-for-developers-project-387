module.exports = {
  ci: {
    collect: {
      staticDistDir: './web',
    },
    upload: {
      target: 'temporary-public-storage',
    },
    reporting: {
      html: {
        directory: './lhci-reports',
      },
    },
  },
};