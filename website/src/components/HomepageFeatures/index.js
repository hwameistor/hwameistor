import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'Localize I/O',
    Svg: require('@site/static/img/feature/01_localized_io.svg').default,
    description: (
      <>
      </>
    ),
  },
  {
    title: 'Backup and Restore',
    Svg: require('@site/static/img/feature/02_backup_and_restore.svg').default,
    description: (
      <>
      </>
    ),
  },
  {
    title: 'Auto Expansion',
    Svg: require('@site/static/img/feature/03_auto_expansion.svg').default,
    description: (
      <>
      </>
    ),
  },
  {
    title: 'Node Affinity',
    Svg: require('@site/static/img/feature/04_node_affinity.svg').default,
    description: (
      <>
      </>
    ),
  },
  {
    title: 'High Availability',
    Svg: require('@site/static/img/feature/05_high_availability.svg').default,
    description: (
      <>
      </>
    ),
  },
  {
    title: 'Application-Aware Scheduling',
    Svg: require('@site/static/img/feature/06_application-aware_scheduling.svg').default,
    description: (
      <>
      </>
    ),
  },
  {
    title: 'Disk Health Management',
    Svg: require('@site/static/img/feature/08_disk_health_management.svg').default,
    description: (
      <>
      </>
    ),
  },
  {
    title: 'Control/Data Plane Separation',
    Svg: require('@site/static/img/feature/07_control_data_plane_separation.svg').default,
    description: (
      <>
      </>
    ),
  },
];

function Feature({Svg, title, description}) {
  return (
    <div className={clsx('col col--3')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
