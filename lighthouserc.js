module.exports = {
  ci: {
    collect: {
      staticDistDir: './web',
    },
    upload: {
      target: 'filesystem',
      outputDir: './lhci-reports',
      reportFilenamePattern: '%%HOSTNAME%%-%%PATHNAME%%-%%DATETIME%%.report.%%EXTENSION%%',
    },
  },
};