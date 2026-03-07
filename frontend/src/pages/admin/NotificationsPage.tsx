import { Button, Card, Col, Form, Input, Row, Select, Space, Table, Tag, Typography, message } from "antd";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import { useCallback, useEffect, useState } from "react";
import {
  listOutboundCalls,
  listStudentContacts,
  retryOutboundCall,
  updateStudentContact,
  type OutboundCall,
  type StudentContact,
} from "@/shared/api/notifications";

type ContactFormValues = {
  student_id: string;
  student_name?: string;
  guardian_name?: string;
  guardian_phone?: string;
  guardian_relation?: string;
};

const statusColorMap: Record<string, string> = {
  requested: "blue",
  connected: "green",
  failed: "red",
  busy: "orange",
  no_answer: "gold",
  cancelled: "default",
};

export function NotificationsPage() {
  const [form] = Form.useForm<ContactFormValues>();
  const [contacts, setContacts] = useState<StudentContact[]>([]);
  const [calls, setCalls] = useState<OutboundCall[]>([]);
  const [loadingContacts, setLoadingContacts] = useState(false);
  const [loadingCalls, setLoadingCalls] = useState(false);
  const [savingContact, setSavingContact] = useState(false);
  const [keyword, setKeyword] = useState("");
  const [contactPage, setContactPage] = useState(1);
  const [callPage, setCallPage] = useState(1);
  const [contactTotal, setContactTotal] = useState(0);
  const [callTotal, setCallTotal] = useState(0);
  const pageSize = 10;

  const loadContacts = useCallback(async (page: number, currentKeyword: string) => {
    setLoadingContacts(true);
    try {
      const result = await listStudentContacts({ page, pageSize, keyword: currentKeyword });
      setContacts(result.items);
      setContactPage(result.page);
      setContactTotal(result.total);
    } finally {
      setLoadingContacts(false);
    }
  }, []);

  const loadCalls = useCallback(async (page: number, currentKeyword: string) => {
    setLoadingCalls(true);
    try {
      const result = await listOutboundCalls({ page, pageSize, keyword: currentKeyword });
      setCalls(result.items);
      setCallPage(result.page);
      setCallTotal(result.total);
    } finally {
      setLoadingCalls(false);
    }
  }, []);

  useEffect(() => {
    void loadContacts(1, "");
    void loadCalls(1, "");
  }, [loadCalls, loadContacts]);

  const handleSearch = async (value: string) => {
    const nextKeyword = value.trim();
    setKeyword(nextKeyword);
    await Promise.all([loadContacts(1, nextKeyword), loadCalls(1, nextKeyword)]);
  };

  const handleEdit = (contact: StudentContact) => {
    form.setFieldsValue({
      student_id: contact.student_id,
      student_name: contact.student_name,
      guardian_name: contact.guardian_name,
      guardian_phone: contact.guardian_phone,
      guardian_relation: contact.guardian_relation,
    });
  };

  const handleSubmit = async (values: ContactFormValues) => {
    setSavingContact(true);
    try {
      await updateStudentContact(values.student_id.trim(), {
        student_name: values.student_name?.trim() ?? "",
        guardian_name: values.guardian_name?.trim() ?? "",
        guardian_phone: values.guardian_phone ?? "",
        guardian_relation: values.guardian_relation?.trim() ?? "",
      });
      message.success("Contact updated");
      await Promise.all([loadContacts(contactPage, keyword), loadCalls(callPage, keyword)]);
    } finally {
      setSavingContact(false);
    }
  };

  const handleRetry = async (id: string) => {
    await retryOutboundCall(id);
    message.success("Outbound call retried");
    await loadCalls(callPage, keyword);
  };

  const contactColumns: ColumnsType<StudentContact> = [
    { title: "Student ID", dataIndex: "student_id" },
    { title: "Student", dataIndex: "student_name" },
    { title: "Guardian", dataIndex: "guardian_name" },
    { title: "Phone", dataIndex: "guardian_phone", render: (value: string) => value || "-" },
    { title: "Relation", dataIndex: "guardian_relation", render: (value: string) => value || "-" },
    {
      title: "Action",
      render: (_, record) => (
        <Button type="link" onClick={() => handleEdit(record)}>
          Edit
        </Button>
      ),
    },
  ];

  const callColumns: ColumnsType<OutboundCall> = [
    { title: "Student", dataIndex: "student_name", render: (value: string, record) => value || record.student_id },
    { title: "Guardian", dataIndex: "guardian_name", render: (value: string) => value || "-" },
    { title: "Phone", dataIndex: "guardian_phone", render: (value: string) => value || "-" },
    {
      title: "Status",
      dataIndex: "status",
      render: (value: string) => <Tag color={statusColorMap[value] ?? "default"}>{value}</Tag>,
    },
    { title: "Provider", dataIndex: "provider" },
    { title: "Trigger", dataIndex: "trigger_source" },
    {
      title: "Requested At",
      dataIndex: "requested_at",
      render: (value: string) => new Date(value).toLocaleString(),
    },
    {
      title: "Action",
      render: (_, record) => (
        <Button type="link" onClick={() => void handleRetry(record.id)}>
          Retry
        </Button>
      ),
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Typography.Title level={3} style={{ marginBottom: 0 }}>
        Smart Outbound Calls
      </Typography.Title>

      <Card>
        <Input.Search
          allowClear
          placeholder="Search by student ID, name, guardian or phone"
          enterButton="Search"
          onSearch={(value) => void handleSearch(value)}
        />
      </Card>

      <Row gutter={[16, 16]}>
        <Col span={10}>
          <Card title="Guardian Contact">
            <Form<ContactFormValues> form={form} layout="vertical" onFinish={(values) => void handleSubmit(values)}>
              <Form.Item label="Student ID" name="student_id" rules={[{ required: true, message: "Please input student ID" }]}>
                <Input placeholder="20260001" />
              </Form.Item>
              <Form.Item label="Student Name" name="student_name">
                <Input placeholder="Student name" />
              </Form.Item>
              <Form.Item label="Guardian Name" name="guardian_name">
                <Input placeholder="Guardian name" />
              </Form.Item>
              <Form.Item label="Guardian Phone" name="guardian_phone">
                <Input placeholder="Leave empty to clear" />
              </Form.Item>
              <Form.Item label="Relation" name="guardian_relation">
                <Select
                  allowClear
                  options={[
                    { label: "Father", value: "father" },
                    { label: "Mother", value: "mother" },
                    { label: "Grandparent", value: "grandparent" },
                    { label: "Other", value: "other" },
                  ]}
                />
              </Form.Item>
              <Space>
                <Button type="primary" htmlType="submit" loading={savingContact}>
                  Save Contact
                </Button>
                <Button onClick={() => form.resetFields()}>Reset</Button>
              </Space>
            </Form>
          </Card>
        </Col>

        <Col span={14}>
          <Card title="Contact List">
            <Table
              rowKey="student_id"
              columns={contactColumns}
              dataSource={contacts}
              loading={loadingContacts}
              pagination={
                {
                  current: contactPage,
                  pageSize,
                  total: contactTotal,
                  onChange: (page: number) => {
                    void loadContacts(page, keyword);
                  },
                } as TablePaginationConfig
              }
            />
          </Card>
        </Col>
      </Row>

      <Card title="Outbound Call Records">
        <Table
          rowKey="id"
          columns={callColumns}
          dataSource={calls}
          loading={loadingCalls}
          pagination={
            {
              current: callPage,
              pageSize,
              total: callTotal,
              onChange: (page: number) => {
                void loadCalls(page, keyword);
              },
            } as TablePaginationConfig
          }
          expandable={{
            expandedRowRender: (record) => (
              <Space direction="vertical" size={4}>
                <Typography.Text>{record.message}</Typography.Text>
                {record.error ? <Typography.Text type="danger">{record.error}</Typography.Text> : null}
              </Space>
            ),
          }}
        />
      </Card>
    </Space>
  );
}
