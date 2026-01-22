import { FC } from 'react';
import { Nav } from 'react-bootstrap';
import { NavLink, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const TabNav: FC<{ menus: { name: string; path: string }[] }> = ({ menus }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'nav_menus' });
  const { pathname } = useLocation();
  return (
    <Nav variant="underline" className="mb-4 border-bottom">
      {menus.map((menu) => (
        <Nav.Item key={menu.path}>
          <NavLink
            to={menu.path}
            className={() =>
              pathname === menu.path ? 'nav-link active' : 'nav-link'
            }>
            {t(menu.name)}
          </NavLink>
        </Nav.Item>
      ))}
    </Nav>
  );
};

export default TabNav;
