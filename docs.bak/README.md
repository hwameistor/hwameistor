# Documentation Website

## URL:

[https://hwameistor.io](https://hwameistor.io/)

## Static Backend

Built using [Docusaurus 2](https://docusaurus.io/), a modern static website generator.

All markdown documents are placed under `./docs`. 

Their translations are placed under `./i18n/cn/docusaurus-plugin-content-docs/current`

## Install Docusaurus modules

```bash
$ npm ci
```

## Local development

```
$ npm start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

Note: You may not view `translation switch` using local development server. To view Chinese translation site **only**:

```bash
$ npm start -- --locale cn
```

## Build

```bash
$ npm run build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

## Host build locally

Note: Only `build` can view `translation switch`

```bash
$ npm run serve
```