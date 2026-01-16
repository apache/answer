import { FC } from 'react';
import './index.scss';

interface BubbleUserProps {
  content?: string;
}

const BubbleUser: FC<BubbleUserProps> = ({ content }) => {
  return (
    <div className="text-end bubble-user-wrap">
      <div className="d-inline-block text-start bubble-user p-3 rounded pre-line">
        {content}
      </div>
    </div>
  );
};

export default BubbleUser;
