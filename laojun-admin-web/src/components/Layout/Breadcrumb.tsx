import { Breadcrumb as AntBreadcrumb } from 'antd';
import { Link } from 'react-router-dom';
import { HomeOutlined } from '@ant-design/icons';
import { useAppStore } from '@/stores/app';

const Breadcrumb: React.FC = () => {
  const { breadcrumbs } = useAppStore();

  if (!breadcrumbs || breadcrumbs.length === 0) {
    return null;
  }

  const items = breadcrumbs.map((item, index) => {
    const isLast = index === breadcrumbs.length - 1;
    
    return {
      key: index,
      title: (
        <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
          {index === 0 && <HomeOutlined />}
          {item.path && !isLast ? (
            <Link to={item.path}>{item.title}</Link>
          ) : (
            <span>{item.title}</span>
          )}
        </span>
      ),
    };
  });

  return (
    <AntBreadcrumb
      items={items}
      style={{
        margin: '0',
        fontSize: '14px',
      }}
    />
  );
};

export default Breadcrumb;