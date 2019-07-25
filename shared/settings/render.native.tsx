import * as React from 'react'
import SettingsNav from './nav'

import {Props} from './render'

const SettingsRender = (props: Props) => {
  const {loadHasRandomPW} = props
  React.useEffect(() => {
    loadHasRandomPW()
  }, [loadHasRandomPW])

  return (
    <SettingsNav
      badgeNotifications={props.badgeNotifications}
      badgeNumbers={props.badgeNumbers}
      logoutInProgress={props.logoutInProgress}
      selectedTab={props.selectedTab}
      onTabChange={props.onTabChange}
      onLogout={props.onLogout}
      hasRandomPW={props.hasRandomPW || null}
    />
  )
}

export default SettingsRender
