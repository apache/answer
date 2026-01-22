import { Modal, Form, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

const Index = ({ visible, api_key = '', onClose }) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.apikeys.created_modal',
  });

  return (
    <Modal show={visible} onHide={onClose}>
      <Modal.Header closeButton>{t('title')}</Modal.Header>
      <Modal.Body>
        <Form>
          <Form.Group controlId="api_key" className="mb-3">
            <Form.Label>{t('api_key')}</Form.Label>
            <Form.Control
              type="text"
              defaultValue={api_key}
              readOnly
              disabled
            />
          </Form.Group>

          <div className="mb-3">{t('description')}</div>
        </Form>
      </Modal.Body>
      <Modal.Footer>
        <Button variant="link" onClick={onClose}>
          {t('close', { keyPrefix: 'btns' })}
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default Index;
