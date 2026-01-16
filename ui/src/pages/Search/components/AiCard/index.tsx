import { useState, useEffect } from 'react';
import { Card, Spinner } from 'react-bootstrap';
import { useSearchParams, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { v4 as uuidv4 } from 'uuid';

import { BubbleAi, BubbleUser } from '@/components';
import { aiControlStore } from '@/stores';
import * as Type from '@/common/interface';
import requestAi from '@/utils/requestAi';

const Index = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'ai_assistant' });
  const { ai_enabled } = aiControlStore((state) => state);
  const [searchParams] = useSearchParams();
  const [isLoading, setIsLoading] = useState(false);
  const [isGenerate, setIsGenerate] = useState(false);
  const [isCompleted, setIsCompleted] = useState(false);
  const [conversions, setConversions] = useState<Type.ConversationDetail>({
    records: [],
    conversation_id: '',
    created_at: 0,
    topic: '',
    updated_at: 0,
  });

  const handleSubmit = async (userMsg) => {
    setIsLoading(true);
    setIsCompleted(false);
    const newConversationId = uuidv4();
    setConversions({
      conversation_id: newConversationId,
      created_at: 0,
      topic: '',
      updated_at: 0,
      records: [
        {
          chat_completion_id: Date.now().toString(),
          role: 'user',
          content: userMsg,
          helpful: 0,
          unhelpful: 0,
          created_at: Date.now(),
        },
      ],
    });

    const params = {
      conversation_id: newConversationId,
      messages: [
        {
          role: 'user',
          content: userMsg,
        },
      ],
    };

    await requestAi('/answer/api/v1/chat/completions', {
      body: JSON.stringify(params),
      onMessage: (res) => {
        if (!res.choices[0].delta?.content) {
          return;
        }
        setIsLoading(false);
        setIsGenerate(true);

        setConversions((prev) => {
          const updatedRecords = [...prev.records];
          const lastConversion = updatedRecords[updatedRecords.length - 1];
          if (lastConversion?.chat_completion_id === res?.chat_completion_id) {
            updatedRecords[updatedRecords.length - 1] = {
              ...lastConversion,
              content: lastConversion.content + res.choices[0].delta.content,
            };
          } else {
            updatedRecords.push({
              chat_completion_id: res.chat_completion_id,
              role: res.choices[0].delta.role || 'assistant',
              content: res.choices[0].delta.content,
              helpful: 0,
              unhelpful: 0,
              created_at: Date.now(),
            });
          }
          return {
            ...prev,
            conversation_id: params.conversation_id,
            records: updatedRecords,
          };
        });
      },
      onError: (error) => {
        setIsGenerate(false);
        setIsLoading(false);
        setIsCompleted(true);
        console.error('Error:', error);
      },
      onComplete: () => {
        setIsCompleted(true);
        setIsGenerate(false);
      },
    });
  };

  useEffect(() => {
    const q = searchParams.get('q') || '';
    if (ai_enabled && q) {
      handleSubmit(q);
    }
  }, [searchParams]);

  if (!ai_enabled) {
    return null;
  }
  return (
    <Card className="mb-5">
      <Card.Header>
        {t('ai_assistant', { keyPrefix: 'page_title' })}
      </Card.Header>
      <Card.Body>
        {conversions?.records.map((item, index) => {
          const isLastMessage =
            index === Number(conversions?.records.length) - 1;
          return (
            <div
              key={`${item.chat_completion_id}-${item.role}`}
              className={`${isLastMessage ? '' : 'mb-4'}`}>
              {item.role === 'user' ? (
                <BubbleUser content={item.content} />
              ) : (
                <BubbleAi
                  canType
                  chatId={item.chat_completion_id}
                  isLast
                  isCompleted={!isGenerate}
                  content={item.content}
                  actionData={{
                    helpful: item.helpful,
                    unhelpful: item.unhelpful,
                  }}
                />
              )}
            </div>
          );
        })}
        {isLoading && (
          <Spinner
            animation="border"
            size="sm"
            variant="secondary"
            className="mt-4"
          />
        )}
      </Card.Body>
      {isCompleted && !isLoading && (
        <Card.Footer className="py-3">
          <Link
            className="btn btn-outline-primary me-3"
            to={`/ai-assistant/${conversions.conversation_id}`}>
            {t('ask_a_follow_up')}
          </Link>
          <span className="small text-secondary">{t('ai_generate')}</span>
        </Card.Footer>
      )}
    </Card>
  );
};

export default Index;
